ALTER TABLE entry ADD COLUMN shown TEXT NOT NULL DEFAULT 'auto' CHECK (shown IN ('auto', 'public', 'private'));
UPDATE entry SET shown = CASE visibility WHEN 'private' THEN 'private' ELSE 'auto' END;
ALTER TABLE entry DROP COLUMN visibility;
ALTER TABLE entry RENAME COLUMN shown TO visibility;

CREATE TABLE setting (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
