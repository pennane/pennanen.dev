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

DROP TABLE claim_reading;

ALTER TABLE claim RENAME TO legacy_claim;
ALTER TABLE observation RENAME TO legacy_observation;
ALTER TABLE relation RENAME TO legacy_relation;
ALTER TABLE part RENAME TO legacy_part;

CREATE VIEW claim AS
  SELECT i.item_id, c.source, c.external_id, c.data, c.fetched_at
  FROM record_claim c
  JOIN item_identity i ON i.source = c.source AND i.external_id = c.external_id;

CREATE VIEW claim_reading AS
  SELECT i.item_id, r.source, r.version, c.fetched_at, r.reading
  FROM record_reading r
  JOIN record_claim c ON c.source = r.source AND c.external_id = r.external_id
  JOIN item_identity i ON i.source = r.source AND i.external_id = r.external_id;

CREATE VIEW observation AS
  SELECT o.id, i.item_id, o.source, o.signal, o.key, o.observed_at, o.data
  FROM record_observation o
  JOIN item_identity i ON i.source = o.about_source AND i.external_id = o.about_external_id;

CREATE VIEW relation AS
  SELECT DISTINCT f.item_id AS from_id, l.type, t.item_id AS to_id
  FROM link l
  JOIN item_identity f ON f.source = l.from_source AND f.external_id = l.from_external_id
  JOIN item_identity t ON t.source = l.to_source AND t.external_id = l.to_external_id
  WHERE l.position IS NULL AND f.item_id <> t.item_id;

CREATE VIEW part AS
  SELECT t.item_id AS whole_id, min(l.position) AS position, f.item_id AS part_id
  FROM link l
  JOIN item_identity f ON f.source = l.from_source AND f.external_id = l.from_external_id
  JOIN item_identity t ON t.source = l.to_source AND t.external_id = l.to_external_id
  WHERE l.type = 'part-of' AND f.item_id <> t.item_id
  GROUP BY t.item_id, f.item_id;
