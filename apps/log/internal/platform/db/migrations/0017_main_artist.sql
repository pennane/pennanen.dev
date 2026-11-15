INSERT OR IGNORE INTO relation (from_id, type, to_id)
SELECT credit.from_id, 'with', credit.to_id
FROM relation credit
JOIN claim album ON album.item_id = credit.from_id AND album.source = 'spotify-album'
JOIN claim artist ON artist.item_id = credit.to_id AND artist.source = 'spotify-artist'
WHERE credit.type = 'by'
  AND json_extract(album.data, '$.artists[0].id') IS NOT NULL
  AND (artist.external_id IS NOT json_extract(album.data, '$.artists[0].id')
       OR artist.external_id = '0LyfQWJT6nXafLPZqxe9Of');

DELETE FROM relation
WHERE type = 'by'
  AND EXISTS (
    SELECT 1 FROM relation other
    WHERE other.from_id = relation.from_id AND other.type = 'with' AND other.to_id = relation.to_id);
