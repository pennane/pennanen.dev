CREATE TABLE take_field_shown (
  item_id INTEGER NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  source  TEXT NOT NULL,
  field   TEXT NOT NULL,
  PRIMARY KEY (item_id, source, field)
);

CREATE TABLE observation (
  id          INTEGER PRIMARY KEY,
  item_id     INTEGER NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  source      TEXT NOT NULL,
  signal      TEXT NOT NULL,
  key         TEXT NOT NULL,
  observed_at TEXT NOT NULL,
  data        TEXT NOT NULL CHECK (json_valid(data)),
  UNIQUE (source, key)
);
CREATE INDEX observation_item ON observation (item_id, source, signal, observed_at);
