package history

import "testing"

func TestHistoryWhereKeepsPHPGuestPlayTimelineBug(t *testing.T) {
	where, args := historyWhere(KindPlay, tableSpec{owner: "sid", timeField: "playtime"}, 0, "guest", 2, 1700000000)
	if where != "sid=? AND showtype=0 AND playtime BETWEEN ? AND ?" {
		t.Fatalf("where = %q", where)
	}
	if args[1] != int64(1699395200) || args[2] != int64(1697408000) {
		t.Fatalf("args = %#v", args)
	}
}

func TestHistoryWhereUserPlayTimelineUsesAscendingBounds(t *testing.T) {
	_, args := historyWhere(KindPlay, tableSpec{owner: "uid", timeField: "playtime"}, 7, "", 2, 1700000000)
	if args[1] != int64(1697408000) || args[2] != int64(1699395200) {
		t.Fatalf("args = %#v", args)
	}
}

func TestHistoryWhereDownTimeline(t *testing.T) {
	where, args := historyWhere(KindDown, tableSpec{owner: "uid", timeField: "downtime"}, 7, "", 3, 1700000000)
	if where != "uid=? AND showtype=0 AND downtime<?" {
		t.Fatalf("where = %q", where)
	}
	if args[1] != int64(1697408000) {
		t.Fatalf("args = %#v", args)
	}
}

func TestMiniPlaySpecUsesPartitionTables(t *testing.T) {
	userSpec := miniPlaySpec(101, "")
	if userSpec.table != "minivod_viewlogs_01" || userSpec.owner != "uid" {
		t.Fatalf("user spec = %#v", userSpec)
	}
	guestSpec := miniPlaySpec(0, "250f")
	if guestSpec.table != "minivod_guestviewlogs_2" || guestSpec.owner != "sid" {
		t.Fatalf("guest spec = %#v", guestSpec)
	}
}

func TestHistoryWhereMiniPlayTimelineDoesNotFilterShowtype(t *testing.T) {
	where, args := historyWhere(KindMiniPlay, tableSpec{owner: "uid", timeField: "playtime"}, 7, "", 2, 1700000000)
	if where != "uid=? AND playtime BETWEEN ? AND ?" {
		t.Fatalf("where = %q", where)
	}
	if args[1] != int64(1697408000) || args[2] != int64(1699395200) {
		t.Fatalf("args = %#v", args)
	}
}
