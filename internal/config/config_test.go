package config

import "testing"

func TestFromEnvReadsOptionalMySQLDatabaseDSNs(t *testing.T) {
	t.Setenv("MYSQL_READ_DSN", "reader:pass@tcp(replica-db:3306)/app")
	t.Setenv("MYSQL_LOG_DSN", "log-user:log-pass@tcp(log-db:3306)/logs")

	cfg := FromEnv()
	if cfg.MySQLReadDSN != "reader:pass@tcp(replica-db:3306)/app" {
		t.Fatalf("MySQLReadDSN=%q", cfg.MySQLReadDSN)
	}
	if cfg.MySQLLogDSN != "log-user:log-pass@tcp(log-db:3306)/logs" {
		t.Fatalf("MySQLLogDSN=%q", cfg.MySQLLogDSN)
	}
}
