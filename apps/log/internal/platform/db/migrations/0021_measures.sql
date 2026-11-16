CREATE TABLE measure (
  item_id INTEGER NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  source  TEXT NOT NULL,
  field   TEXT NOT NULL,
  value   INTEGER NOT NULL,
  PRIMARY KEY (item_id, field)
);

CREATE INDEX measure_source ON measure (source);

CREATE TEMP TABLE film_of_item AS
  SELECT l.item_id, l.external_id AS film_id
  FROM item_identity l
  WHERE l.source = 'letterboxd'
    AND l.id = (SELECT min(id) FROM item_identity WHERE item_id = l.item_id AND source = 'letterboxd');

CREATE TEMP TABLE take_key AS
  SELECT t.item_id, t.external_id AS old_external_id, 'letterboxd:' || f.film_id AS new_external_id, c.changed_at, c.fetched_at
  FROM item_identity t
  JOIN record_claim c ON c.source = t.source AND c.external_id = t.external_id
  JOIN film_of_item f ON f.item_id = t.item_id
  WHERE t.source = 'letterboxd-take';

INSERT INTO take_key (item_id, old_external_id, new_external_id, changed_at, fetched_at)
  SELECT k.item_id, c.external_id, c.external_id, c.changed_at, c.fetched_at
  FROM record_claim c
  JOIN (SELECT DISTINCT item_id, new_external_id FROM take_key) k ON k.new_external_id = c.external_id
  WHERE c.source = 'letterboxd-take'
    AND c.external_id NOT IN (SELECT old_external_id FROM take_key);

CREATE TEMP TABLE take_superseded AS
  SELECT old_external_id
  FROM (
    SELECT old_external_id,
           ROW_NUMBER() OVER (
             PARTITION BY new_external_id
             ORDER BY changed_at DESC, fetched_at DESC, old_external_id = new_external_id DESC, old_external_id
           ) AS newness
    FROM take_key
  )
  WHERE newness > 1;

CREATE TEMP TABLE take_move AS
  SELECT old_external_id, new_external_id
  FROM take_key
  WHERE old_external_id <> new_external_id
    AND old_external_id NOT IN (SELECT old_external_id FROM take_superseded);

DELETE FROM item_identity
  WHERE source = 'letterboxd-take' AND external_id IN (SELECT old_external_id FROM take_superseded);

DELETE FROM item_overlap
  WHERE source = 'letterboxd-take' AND external_id IN (SELECT old_external_id FROM take_superseded);

DELETE FROM identity_waiting
  WHERE (waiting_source = 'letterboxd-take' AND waiting_external_id IN (SELECT old_external_id FROM take_superseded))
     OR (about_source = 'letterboxd-take' AND about_external_id IN (SELECT old_external_id FROM take_superseded));

DELETE FROM record_claim
  WHERE source = 'letterboxd-take' AND external_id IN (SELECT old_external_id FROM take_superseded);

INSERT INTO record_claim (source, external_id, data, fetched_at, changed_at)
  SELECT c.source, m.new_external_id, c.data, c.fetched_at, c.changed_at
  FROM record_claim c
  JOIN take_move m ON m.old_external_id = c.external_id
  WHERE c.source = 'letterboxd-take';

INSERT INTO record_reading (source, external_id, version, reading)
  SELECT r.source, m.new_external_id, r.version, r.reading
  FROM record_reading r
  JOIN take_move m ON m.old_external_id = r.external_id
  WHERE r.source = 'letterboxd-take';

INSERT INTO link (asserted_source, asserted_external_id, from_source, from_external_id, type, to_source, to_external_id, position)
  SELECT l.asserted_source, m.new_external_id, l.from_source, l.from_external_id, l.type, l.to_source, l.to_external_id, l.position
  FROM link l
  JOIN take_move m ON m.old_external_id = l.asserted_external_id
  WHERE l.asserted_source = 'letterboxd-take';

UPDATE link
  SET from_external_id = (SELECT new_external_id FROM take_move WHERE old_external_id = link.from_external_id)
  WHERE from_source = 'letterboxd-take' AND from_external_id IN (SELECT old_external_id FROM take_move);

UPDATE link
  SET to_external_id = (SELECT new_external_id FROM take_move WHERE old_external_id = link.to_external_id)
  WHERE to_source = 'letterboxd-take' AND to_external_id IN (SELECT old_external_id FROM take_move);

UPDATE item_identity
  SET external_id = (SELECT new_external_id FROM take_move WHERE old_external_id = item_identity.external_id)
  WHERE source = 'letterboxd-take' AND external_id IN (SELECT old_external_id FROM take_move);

UPDATE item_overlap
  SET external_id = (SELECT new_external_id FROM take_move WHERE old_external_id = item_overlap.external_id)
  WHERE source = 'letterboxd-take' AND external_id IN (SELECT old_external_id FROM take_move);

UPDATE identity_waiting
  SET waiting_external_id = (SELECT new_external_id FROM take_move WHERE old_external_id = identity_waiting.waiting_external_id)
  WHERE waiting_source = 'letterboxd-take' AND waiting_external_id IN (SELECT old_external_id FROM take_move);

UPDATE identity_waiting
  SET about_external_id = (SELECT new_external_id FROM take_move WHERE old_external_id = identity_waiting.about_external_id)
  WHERE about_source = 'letterboxd-take' AND about_external_id IN (SELECT old_external_id FROM take_move);

UPDATE record_observation
  SET about_external_id = (SELECT new_external_id FROM take_move WHERE old_external_id = record_observation.about_external_id)
  WHERE about_source = 'letterboxd-take' AND about_external_id IN (SELECT old_external_id FROM take_move);

DELETE FROM record_claim
  WHERE source = 'letterboxd-take' AND external_id IN (SELECT old_external_id FROM take_move);

CREATE TEMP TABLE item_keyed_watch AS
  SELECT o.id, f.film_id
  FROM record_observation o
  JOIN item_identity a ON a.source = o.about_source AND a.external_id = o.about_external_id
  JOIN film_of_item f ON f.item_id = a.item_id
  WHERE o.source = 'letterboxd'
    AND o.signal = 'watched'
    AND substr(o.key, 1, 5) = 'item/'
    AND instr(o.key, '/watched/') > 0;

UPDATE OR IGNORE record_observation
  SET key = 'letterboxd:' || (SELECT film_id FROM item_keyed_watch w WHERE w.id = record_observation.id) || substr(key, instr(key, '/watched/')),
      about_source = 'letterboxd',
      about_external_id = (SELECT film_id FROM item_keyed_watch w WHERE w.id = record_observation.id)
  WHERE id IN (SELECT id FROM item_keyed_watch);

INSERT INTO measure (item_id, source, field, value)
  SELECT item_id, 'spotify', 'listening', listening
  FROM (
    SELECT i.item_id,
           CAST(round(json_extract(c.data, '$.listening')) AS INTEGER) AS listening,
           ROW_NUMBER() OVER (PARTITION BY i.item_id ORDER BY c.fetched_at DESC, c.external_id) AS newness
    FROM record_claim c
    JOIN item_identity i ON i.source = c.source AND i.external_id = c.external_id
    WHERE c.source = 'spotify-listening'
      AND json_type(c.data, '$.listening') IN ('integer', 'real')
  )
  WHERE newness = 1;

UPDATE item_verdict
  SET own = json_remove(own, '$.values.listening', '$.sourceOf.listening')
  WHERE json_extract(own, '$.sourceOf.listening') = 'spotify-listening';

UPDATE item_verdict
  SET verdicts = json_remove(verdicts, '$.values.listening', '$.sourceOf.listening')
  WHERE json_extract(verdicts, '$.sourceOf.listening') = 'spotify-listening';

UPDATE setting
  SET value = (
    SELECT json_group_array(
      CASE WHEN json_extract(axis.value, '$.reading.source') = 'spotify-listening'
        THEN json_set(axis.value, '$.reading.source', 'spotify')
        ELSE json(axis.value)
      END)
    FROM (SELECT key, value FROM json_each(setting.value) ORDER BY key) axis)
  WHERE key = 'axes' AND json_valid(value) AND json_type(value) = 'array';

DELETE FROM item_identity WHERE source = 'spotify-listening';

DELETE FROM item_overlap WHERE source = 'spotify-listening';

DELETE FROM identity_waiting WHERE waiting_source = 'spotify-listening' OR about_source = 'spotify-listening';

DELETE FROM record_claim WHERE source = 'spotify-listening';

DROP TABLE item_keyed_watch;

DROP TABLE take_move;

DROP TABLE take_superseded;

DROP TABLE take_key;

DROP TABLE film_of_item;
