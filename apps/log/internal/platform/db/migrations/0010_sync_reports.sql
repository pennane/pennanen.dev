CREATE TABLE sync_report (
  name   TEXT PRIMARY KEY,
  report TEXT NOT NULL CHECK (json_valid(report))
);
