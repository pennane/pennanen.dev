CREATE TABLE record_claim (
  source      TEXT NOT NULL,
  external_id TEXT NOT NULL,
  data        TEXT NOT NULL CHECK (json_valid(data)),
  fetched_at  TEXT NOT NULL,
  changed_at  TEXT NOT NULL,
  PRIMARY KEY (source, external_id)
);

CREATE TABLE record_reading (
  source      TEXT NOT NULL,
  external_id TEXT NOT NULL,
  version     INTEGER NOT NULL,
  reading     TEXT NOT NULL CHECK (json_valid(reading)),
  PRIMARY KEY (source, external_id),
  FOREIGN KEY (source, external_id) REFERENCES record_claim (source, external_id) ON DELETE CASCADE
);

CREATE TABLE record_observation (
  id               INTEGER PRIMARY KEY,
  about_source     TEXT NOT NULL,
  about_external_id TEXT NOT NULL,
  source           TEXT NOT NULL,
  signal           TEXT NOT NULL,
  key              TEXT NOT NULL,
  observed_at      TEXT NOT NULL,
  data             TEXT NOT NULL CHECK (json_valid(data)),
  UNIQUE (source, key)
);

CREATE INDEX record_observation_about ON record_observation (about_source, about_external_id, source, signal, observed_at);
