UPDATE score SET axis = 'prestige' WHERE axis = 'enjoyment';

UPDATE score SET division = CASE division
  WHEN 'loved' THEN 's' WHEN 'masterpiece' THEN 's'
  WHEN 'liked' THEN 'a' WHEN 'great' THEN 'a'
  WHEN 'fine' THEN 'b' WHEN 'solid' THEN 'b'
  WHEN 'meh' THEN 'c' WHEN 'rough' THEN 'c'
  WHEN 'disliked' THEN 'd' WHEN 'broken' THEN 'd'
  ELSE division
END;

INSERT INTO setting (key, value)
SELECT 'axes', json_array(
  json_object('key', 'prestige', 'name', 'prestige', 'divisions', json(tiers.value)),
  json_object(
    'key', 'quality',
    'name', coalesce((SELECT json_extract(value, '$.axisLabels.quality') FROM setting WHERE key = 'kind:game'), 'quality'),
    'divisions', json(tiers.value)
  )
)
FROM (SELECT '[{"key":"s","name":"S"},{"key":"a","name":"A"},{"key":"b","name":"B"},{"key":"c","name":"C"},{"key":"d","name":"D"},{"key":"f","name":"F"}]' AS value) AS tiers
WHERE true
ON CONFLICT (key) DO UPDATE SET value = excluded.value;

UPDATE setting
SET value = json_set(json_remove(value, '$.axisLabels', '$.divisions'), '$.axes', json('["prestige","quality"]'))
WHERE key LIKE 'kind:%';
