DROP INDEX entry_status;

ALTER TABLE entry ADD COLUMN logged_by TEXT NOT NULL DEFAULT 'ap';

UPDATE entry SET logged_by = status_set_by;

CREATE TABLE status_choice (
  item_id   INTEGER PRIMARY KEY REFERENCES item (id) ON DELETE CASCADE,
  status    TEXT NOT NULL,
  chosen_at TEXT NOT NULL
);

CREATE TABLE tag_choice (
  item_id   INTEGER NOT NULL REFERENCES item (id) ON DELETE CASCADE,
  tag       TEXT NOT NULL CHECK (length(tag) > 0),
  presence  TEXT NOT NULL CHECK (presence IN ('added', 'removed')),
  chosen_at TEXT NOT NULL,
  PRIMARY KEY (item_id, tag)
);

CREATE TABLE dismissal (
  item_id      INTEGER PRIMARY KEY REFERENCES item (id) ON DELETE CASCADE,
  dismissed_at TEXT NOT NULL
);

CREATE TABLE proposal (
  item_id    INTEGER NOT NULL REFERENCES item (id) ON DELETE CASCADE,
  source     TEXT NOT NULL,
  rule       TEXT NOT NULL,
  subject    TEXT NOT NULL CHECK (subject IN ('status', 'tag', 'dismissal')),
  value      TEXT NOT NULL,
  strength   TEXT NOT NULL CHECK (strength IN ('applying', 'suggesting')),
  changed_at TEXT,
  PRIMARY KEY (item_id, rule, subject, value)
);

CREATE INDEX proposal_source ON proposal (source);
CREATE INDEX proposal_item_subject ON proposal (item_id, subject);

CREATE TABLE kind_status (
  kind   TEXT NOT NULL,
  status TEXT NOT NULL,
  PRIMARY KEY (kind, status)
);

CREATE VIEW status_word AS
  SELECT c.item_id, c.status, 'ap' AS said_by, c.chosen_at AS said_at
  FROM status_choice c
  JOIN item i ON i.id = c.item_id
  JOIN kind_status k ON k.kind = i.kind AND k.status = c.status
  UNION ALL
  SELECT p.item_id, p.value AS status, p.rule AS said_by, p.changed_at AS said_at
  FROM proposal p
  JOIN item i ON i.id = p.item_id
  JOIN kind_status k ON k.kind = i.kind AND k.status = p.value
  WHERE p.subject = 'status' AND p.strength = 'applying';

CREATE VIEW entry_status AS
  SELECT e.item_id,
         coalesce(w.status, 'untriaged') AS status,
         coalesce(w.said_by, 'nobody') AS decided_by,
         w.said_at AS decided_at
  FROM entry e
  LEFT JOIN (
    SELECT item_id, status, said_by, said_at,
           ROW_NUMBER() OVER (PARTITION BY item_id ORDER BY coalesce(said_at, '') DESC, said_by = 'ap' DESC, said_by) AS newness
    FROM status_word
  ) w ON w.item_id = e.item_id AND w.newness = 1;

CREATE VIEW tag_word AS
  SELECT item_id, tag, presence, 'ap' AS said_by, chosen_at AS said_at
  FROM tag_choice
  UNION ALL
  SELECT item_id, value AS tag, 'added' AS presence, rule AS said_by, changed_at AS said_at
  FROM proposal
  WHERE subject = 'tag' AND strength = 'applying';

CREATE VIEW entry_tag AS
  SELECT w.item_id, w.tag, w.said_by AS decided_by
  FROM (
    SELECT item_id, tag, presence, said_by,
           ROW_NUMBER() OVER (PARTITION BY item_id, tag ORDER BY coalesce(said_at, '') DESC, said_by = 'ap' DESC, said_by) AS newness
    FROM tag_word
  ) w
  JOIN entry e ON e.item_id = w.item_id
  WHERE w.newness = 1 AND w.presence = 'added';

CREATE VIEW shown_entry AS
  SELECT e.*
  FROM entry e
  WHERE NOT EXISTS (SELECT 1 FROM dismissal d WHERE d.item_id = e.item_id);

INSERT INTO status_choice (item_id, status, chosen_at)
  SELECT item_id, status, strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
  FROM entry
  WHERE status_set_by = 'ap' AND status NOT IN ('untriaged', 'dismissed');

INSERT INTO dismissal (item_id, dismissed_at)
  SELECT item_id, updated_at
  FROM entry
  WHERE status = 'dismissed';

INSERT INTO tag_choice (item_id, tag, presence, chosen_at)
  SELECT item_id, tag, 'added', strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
  FROM tag
  WHERE added_by = 'ap';

CREATE TEMP TABLE watched_shelf AS
  SELECT e.item_id,
         coalesce(
           (SELECT c.external_id
            FROM item_identity i
            JOIN record_claim c ON c.source = i.source AND c.external_id = i.external_id
            WHERE i.item_id = e.item_id AND i.source = 'letterboxd-take'
            ORDER BY c.fetched_at DESC, c.external_id
            LIMIT 1),
           k.source || ':' || k.external_id
         ) AS take_external_id,
         substr(e.updated_at, 1, 10) AS since,
         e.updated_at
  FROM entry e
  JOIN item_identity k ON k.id = (SELECT min(id) FROM item_identity WHERE item_id = e.item_id)
  WHERE e.status_set_by = 'letterboxd' AND e.status = 'watched';

UPDATE record_claim
  SET data = json_set(data, '$.onWatchedShelfSince', (
    SELECT since FROM watched_shelf WHERE take_external_id = record_claim.external_id
  ))
  WHERE source = 'letterboxd-take'
    AND external_id IN (SELECT take_external_id FROM watched_shelf);

INSERT INTO record_claim (source, external_id, data, fetched_at, changed_at)
  SELECT 'letterboxd-take', take_external_id, json_object('onWatchedShelfSince', since), updated_at, updated_at
  FROM watched_shelf
  WHERE NOT EXISTS (
    SELECT 1 FROM record_claim c WHERE c.source = 'letterboxd-take' AND c.external_id = take_external_id
  );

INSERT OR IGNORE INTO item_identity (item_id, source, external_id)
  SELECT item_id, 'letterboxd-take', take_external_id
  FROM watched_shelf;

DELETE FROM record_reading
  WHERE source = 'letterboxd-take'
    AND external_id IN (SELECT take_external_id FROM watched_shelf);

DROP TABLE watched_shelf;

ALTER TABLE entry DROP COLUMN status;

ALTER TABLE entry DROP COLUMN status_set_by;

DROP TABLE tag;
