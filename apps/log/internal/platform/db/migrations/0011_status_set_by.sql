ALTER TABLE entry ADD COLUMN status_set_by TEXT NOT NULL DEFAULT 'ap';

UPDATE entry SET status_set_by = coalesce((
  SELECT source FROM claim
  WHERE claim.item_id = entry.item_id AND source <> 'ap'
  ORDER BY fetched_at, rowid
  LIMIT 1
), 'ap')
WHERE status = 'untriaged';
