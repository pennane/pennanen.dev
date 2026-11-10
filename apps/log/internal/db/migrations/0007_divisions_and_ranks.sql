DROP TABLE placement;
DROP TABLE ranking_tier;
DROP TABLE ranking;
DROP TABLE moment_score;

CREATE TABLE division_score (
  item_id  INTEGER NOT NULL REFERENCES entry(item_id) ON DELETE CASCADE,
  axis     TEXT NOT NULL,
  division TEXT NOT NULL CHECK (length(division) > 0),
  position INTEGER CHECK (position >= 0),
  PRIMARY KEY (item_id, axis)
);

INSERT INTO division_score (item_id, axis, division, position)
SELECT
  s.item_id,
  s.axis,
  CASE
    WHEN s.axis = 'enjoyment' THEN
      CASE WHEN s.value >= 86 THEN 'loved' WHEN s.value >= 70 THEN 'liked' WHEN s.value >= 48 THEN 'meh' ELSE 'disliked' END
    WHEN s.axis = 'quality' THEN
      CASE WHEN s.value >= 86 THEN 'masterpiece' WHEN s.value >= 70 THEN 'great' WHEN s.value >= 48 THEN 'rough' ELSE 'broken' END
    ELSE
      CASE WHEN s.value >= 67 THEN 'good' WHEN s.value >= 34 THEN 'fine' ELSE 'bad' END
  END,
  NULL
FROM score s
WHERE s.item_id NOT IN (SELECT item_id FROM entry WHERE status = 'dismissed');

UPDATE division_score SET position = (
  SELECT seeded.position FROM (
    SELECT
      d.item_id,
      ROW_NUMBER() OVER (
        PARTITION BY i.kind, d.division
        ORDER BY s.value DESC, coalesce(json_extract(c.data, '$.playtime_forever'), 0) DESC, i.slug
      ) - 1 AS position
    FROM division_score d
    JOIN item i ON i.id = d.item_id
    JOIN score s ON s.item_id = d.item_id AND s.axis = d.axis
    LEFT JOIN claim c ON c.item_id = d.item_id AND c.source = 'steam'
    WHERE d.axis = 'enjoyment'
  ) seeded
  WHERE seeded.item_id = division_score.item_id
)
WHERE axis = 'enjoyment';

ALTER TABLE score RENAME TO number_score_archive;
ALTER TABLE division_score RENAME TO score;
CREATE INDEX score_division ON score (axis, division, position);
