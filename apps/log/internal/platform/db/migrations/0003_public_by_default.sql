INSERT INTO tag (item_id, tag) SELECT item_id, 'idled' FROM entry WHERE status = 'idled' ON CONFLICT DO NOTHING;
UPDATE entry SET status = 'never-played' WHERE status = 'idled';
UPDATE entry SET visibility = 'public';
