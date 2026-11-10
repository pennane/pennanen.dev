CREATE TABLE ranking (
  id          INTEGER PRIMARY KEY,
  slug        TEXT NOT NULL UNIQUE,
  name        TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  layout      TEXT NOT NULL DEFAULT 'numbered' CHECK (layout IN ('numbered', 'tiered')),
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE ranking_tier (
  ranking_id INTEGER NOT NULL REFERENCES ranking(id) ON DELETE CASCADE,
  position   INTEGER NOT NULL,
  name       TEXT NOT NULL,
  PRIMARY KEY (ranking_id, position)
);

CREATE TABLE placement (
  ranking_id    INTEGER NOT NULL REFERENCES ranking(id) ON DELETE CASCADE,
  item_id       INTEGER NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  position      INTEGER NOT NULL,
  tier_position INTEGER,
  note          TEXT NOT NULL DEFAULT '',
  PRIMARY KEY (ranking_id, item_id),
  UNIQUE (ranking_id, position),
  FOREIGN KEY (ranking_id, tier_position) REFERENCES ranking_tier(ranking_id, position) ON DELETE CASCADE
);

CREATE INDEX placement_item ON placement(item_id);
