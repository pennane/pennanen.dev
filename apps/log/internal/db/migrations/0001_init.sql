CREATE TABLE item (
  id         INTEGER PRIMARY KEY,
  kind       TEXT NOT NULL,
  slug       TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  UNIQUE (kind, slug)
);

CREATE TABLE claim (
  item_id     INTEGER NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  source      TEXT NOT NULL,
  external_id TEXT,
  data        TEXT NOT NULL CHECK (json_valid(data)),
  fetched_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  PRIMARY KEY (item_id, source)
);
CREATE UNIQUE INDEX claim_identity ON claim (source, external_id) WHERE external_id IS NOT NULL;

CREATE TABLE relation (
  from_id INTEGER NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  type    TEXT NOT NULL,
  to_id   INTEGER NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  PRIMARY KEY (from_id, type, to_id),
  CHECK (from_id <> to_id)
);
CREATE INDEX relation_to ON relation (to_id);

CREATE TABLE entry (
  item_id    INTEGER PRIMARY KEY REFERENCES item(id) ON DELETE CASCADE,
  status     TEXT NOT NULL DEFAULT 'untriaged',
  visibility TEXT NOT NULL DEFAULT 'private' CHECK (visibility IN ('private', 'public')),
  notes      TEXT NOT NULL DEFAULT '',
  blurb      TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
CREATE INDEX entry_status ON entry (status);

CREATE TABLE score (
  item_id INTEGER NOT NULL REFERENCES entry(item_id) ON DELETE CASCADE,
  axis    TEXT NOT NULL,
  value   INTEGER NOT NULL CHECK (value BETWEEN 0 AND 100),
  PRIMARY KEY (item_id, axis)
);

CREATE TABLE tag (
  item_id INTEGER NOT NULL REFERENCES entry(item_id) ON DELETE CASCADE,
  tag     TEXT NOT NULL CHECK (length(tag) > 0),
  PRIMARY KEY (item_id, tag)
);
CREATE INDEX tag_tag ON tag (tag);

CREATE TABLE moment (
  id         INTEGER PRIMARY KEY,
  item_id    INTEGER NOT NULL REFERENCES entry(item_id) ON DELETE CASCADE,
  at         TEXT NOT NULL,
  note       TEXT NOT NULL DEFAULT '',
  visibility TEXT NOT NULL DEFAULT 'inherit' CHECK (visibility IN ('inherit', 'private')),
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
CREATE INDEX moment_item ON moment (item_id, at);

CREATE TABLE moment_score (
  moment_id INTEGER NOT NULL REFERENCES moment(id) ON DELETE CASCADE,
  axis      TEXT NOT NULL,
  value     INTEGER NOT NULL CHECK (value BETWEEN 0 AND 100),
  PRIMARY KEY (moment_id, axis)
);

CREATE TABLE moment_ref (
  moment_id INTEGER NOT NULL REFERENCES moment(id) ON DELETE CASCADE,
  item_id   INTEGER NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  PRIMARY KEY (moment_id, item_id)
);
