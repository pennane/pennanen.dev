ALTER TABLE entry ADD COLUMN prose TEXT NOT NULL DEFAULT '';

CREATE TABLE moment_archive AS SELECT id, item_id, at, note, visibility, created_at FROM moment;
CREATE TABLE moment_ref_archive AS SELECT moment_id, item_id FROM moment_ref;

UPDATE entry SET prose = (
  SELECT group_concat('**' || at || '** ' || note, char(10) || char(10) ORDER BY at DESC, id DESC)
  FROM moment
  WHERE moment.item_id = entry.item_id AND visibility = 'inherit' AND trim(note) <> ''
)
WHERE item_id IN (SELECT item_id FROM moment WHERE visibility = 'inherit' AND trim(note) <> '');

DROP TABLE moment_ref;
DROP TABLE moment;
