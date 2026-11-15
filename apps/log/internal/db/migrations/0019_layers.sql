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

CREATE TABLE item_identity (
  id          INTEGER PRIMARY KEY,
  item_id     INTEGER NOT NULL REFERENCES item (id) ON DELETE CASCADE,
  source      TEXT NOT NULL,
  external_id TEXT NOT NULL,
  UNIQUE (source, external_id)
);

CREATE INDEX item_identity_item ON item_identity (item_id);

CREATE TABLE link (
  asserted_source      TEXT NOT NULL,
  asserted_external_id TEXT NOT NULL,
  from_source          TEXT NOT NULL,
  from_external_id     TEXT NOT NULL,
  type                 TEXT NOT NULL,
  to_source            TEXT NOT NULL,
  to_external_id       TEXT NOT NULL,
  position             INTEGER CHECK (position IS NULL OR position >= 1),
  PRIMARY KEY (asserted_source, asserted_external_id, from_source, from_external_id, type, to_source, to_external_id),
  FOREIGN KEY (asserted_source, asserted_external_id) REFERENCES record_claim (source, external_id) ON DELETE CASCADE
);

CREATE INDEX link_to ON link (to_source, to_external_id);
CREATE INDEX link_from ON link (from_source, from_external_id);

CREATE TABLE item_verdict (
  item_id  INTEGER PRIMARY KEY REFERENCES item (id) ON DELETE CASCADE,
  own      TEXT NOT NULL CHECK (json_valid(own)),
  verdicts TEXT NOT NULL CHECK (json_valid(verdicts))
);

CREATE TABLE item_overlap (
  source        TEXT NOT NULL,
  external_id   TEXT NOT NULL,
  item_id       INTEGER NOT NULL REFERENCES item (id) ON DELETE CASCADE,
  other_item_id INTEGER NOT NULL REFERENCES item (id) ON DELETE CASCADE,
  PRIMARY KEY (source, external_id, other_item_id)
);

CREATE TABLE identity_waiting (
  about_source        TEXT NOT NULL,
  about_external_id   TEXT NOT NULL,
  waiting_source      TEXT NOT NULL,
  waiting_external_id TEXT NOT NULL,
  PRIMARY KEY (about_source, about_external_id, waiting_source, waiting_external_id)
);

CREATE INDEX identity_waiting_waiting ON identity_waiting (waiting_source, waiting_external_id);
