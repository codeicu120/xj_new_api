package minivod

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryDatabaseRolesDefaultAndCanBeOverridden(t *testing.T) {
	primary := &sql.DB{}
	replica := &sql.DB{}
	logs := &sql.DB{}
	repo := NewRepository(primary)
	if repo.replicaDB != primary {
		t.Fatal("replica DB should default to the primary DB")
	}
	if repo.logDB != primary {
		t.Fatal("log DB should default to the primary DB")
	}
	repo.WithReplicaDB(replica).WithLogDB(logs)
	if repo.replicaDB != replica || repo.logDB != logs || repo.db != primary {
		t.Fatal("database role overrides must not replace the primary DB")
	}
}

func TestReqListMainDataReadsUseReplicaDB(t *testing.T) {
	primaryDB, primaryMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer primaryDB.Close()
	replicaDB, replicaMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer replicaDB.Close()

	repo := NewRepository(primaryDB).WithReplicaDB(replicaDB)
	query := "SELECT * FROM vods WHERE vodid IN(9) AND showtype=1 ORDER BY vodid DESC"
	replicaMock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"vodid"}).AddRow(9))
	rows, err := repo.VODsByIDs(context.Background(), []int{9}, false)
	if err != nil || len(rows) != 1 {
		t.Fatalf("VODsByIDs rows=%#v err=%v", rows, err)
	}
	if err := replicaMock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if err := primaryMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("primary DB was unexpectedly used: %v", err)
	}
}

func TestPendingAndMarkViewLogsUseLogDB(t *testing.T) {
	primaryDB, primaryMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer primaryDB.Close()
	logDB, logMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer logDB.Close()

	repo := NewRepository(primaryDB).WithLogDB(logDB)
	pendingQuery := "SELECT * FROM minivod_guestviewlogs_a WHERE sid=? AND showtype=0 ORDER BY logid DESC LIMIT ?"
	logMock.ExpectQuery(regexp.QuoteMeta(pendingQuery)).WithArgs("abcdef", 10).
		WillReturnRows(sqlmock.NewRows([]string{"logid", "vodid"}).AddRow(1, 9))
	rows, err := repo.PendingViewLogs(context.Background(), 0, "abcdef", 10)
	if err != nil || len(rows) != 1 {
		t.Fatalf("PendingViewLogs rows=%#v err=%v", rows, err)
	}

	markQuery := "UPDATE minivod_guestviewlogs_a SET reqtime=?, showtype=1 WHERE sid=? AND logid IN(?)"
	logMock.ExpectExec(regexp.QuoteMeta(markQuery)).WithArgs(int64(1700000000), "abcdef", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.MarkViewLogsShown(context.Background(), 0, "abcdef", []int{1}, 1700000000); err != nil {
		t.Fatalf("MarkViewLogsShown: %v", err)
	}

	if err := logMock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if err := primaryMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("primary DB was unexpectedly used: %v", err)
	}
}

func TestMiniViewLogAndCountUseLogDB(t *testing.T) {
	primaryDB, primaryMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer primaryDB.Close()
	logDB, logMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer logDB.Close()

	repo := NewRepository(primaryDB).WithLogDB(logDB)
	viewQuery := "SELECT * FROM minivod_guestviewlogs_a WHERE sid=? AND vodid=? LIMIT 1"
	logMock.ExpectQuery(regexp.QuoteMeta(viewQuery)).WithArgs("abcdef", 9).
		WillReturnRows(sqlmock.NewRows([]string{"logid", "vodid"}).AddRow(1, 9))
	if _, err := repo.MiniViewLog(context.Background(), 0, "abcdef", 9); err != nil {
		t.Fatal(err)
	}
	countQuery := "SELECT COUNT(*) FROM minivod_guestviewlogs_a WHERE sid=? AND showtype=1 AND playtime>=?"
	logMock.ExpectQuery(regexp.QuoteMeta(countQuery)).WithArgs("abcdef", int64(1700000000)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	if _, err := repo.CountMiniViewLogsSince(context.Background(), 0, "abcdef", 1700000000, 1); err != nil {
		t.Fatal(err)
	}
	if err := logMock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if err := primaryMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("primary DB was unexpectedly used: %v", err)
	}
}

func TestRecordMiniMediaSplitsPrimaryCounterAndLogWrite(t *testing.T) {
	primaryDB, primaryMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer primaryDB.Close()
	logDB, logMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer logDB.Close()

	repo := NewRepository(primaryDB).WithLogDB(logDB)
	primaryMock.ExpectBegin()
	primaryMock.ExpectQuery(regexp.QuoteMeta("SELECT playcount_lasttime FROM vods WHERE vodid=?")).WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"playcount_lasttime"}).AddRow(0))
	primaryMock.ExpectExec("UPDATE vods SET .* WHERE vodid=\\?").WithArgs(int64(1700000000), 9).
		WillReturnResult(sqlmock.NewResult(0, 1))
	primaryMock.ExpectCommit()

	logMock.ExpectBegin()
	logMock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM minivod_guestviewlogs_a WHERE sid=? AND vodid=? LIMIT 1 FOR UPDATE")).
		WithArgs("abcdef", 9).WillReturnRows(sqlmock.NewRows([]string{"logid"}))
	logMock.ExpectExec(regexp.QuoteMeta("INSERT INTO minivod_guestviewlogs_a(sid, vodid, playtime, deduct, reqtime, showtype) VALUES(?, ?, ?, ?, ?, 1)")).
		WithArgs("abcdef", 9, int64(1700000000), 0, int64(1700000000)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	logMock.ExpectCommit()

	if err := repo.RecordMiniMedia(context.Background(), 0, "abcdef", 9, true, 0, 1700000000); err != nil {
		t.Fatalf("RecordMiniMedia: %v", err)
	}
	if err := primaryMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("primary expectations: %v", err)
	}
	if err := logMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("log expectations: %v", err)
	}
}

func TestPullCandidateReadsUsePrimaryDB(t *testing.T) {
	primaryDB, primaryMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer primaryDB.Close()
	replicaDB, replicaMock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer replicaDB.Close()

	repo := NewRepository(primaryDB).WithReplicaDB(replicaDB)
	query := "SELECT vodid FROM vods WHERE authorid IN(7) AND showtype=1 AND isvip=0 ORDER BY vodid DESC LIMIT 500"
	primaryMock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"vodid"}).AddRow(9))
	ids, err := repo.vodIDsByAuthors(context.Background(), []int{7})
	if err != nil || len(ids) != 1 || ids[0] != 9 {
		t.Fatalf("vodIDsByAuthors ids=%v err=%v", ids, err)
	}
	if err := primaryMock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if err := replicaMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("replica DB was unexpectedly used: %v", err)
	}
}

func TestLongToShortMapByLongIDNilDB(t *testing.T) {
	repo := NewRepository(nil)

	row, err := repo.LongToShortMapByLongID(context.Background(), 9)
	if err != nil {
		t.Fatalf("LongToShortMapByLongID: %v", err)
	}
	if len(row) != 0 {
		t.Fatalf("expected empty row, got %#v", row)
	}
}
