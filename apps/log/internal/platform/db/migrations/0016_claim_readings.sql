CREATE TABLE claim_reading (
  item_id  INTEGER NOT NULL,
  source   TEXT NOT NULL,
  version  INTEGER NOT NULL,
  fetched_at TEXT NOT NULL,
  reading  TEXT NOT NULL CHECK (json_valid(reading)),
  PRIMARY KEY (item_id, source),
  FOREIGN KEY (item_id, source) REFERENCES claim (item_id, source) ON DELETE CASCADE
);
