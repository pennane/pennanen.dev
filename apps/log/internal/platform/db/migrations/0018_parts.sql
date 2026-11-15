CREATE TABLE part (
  whole_id INTEGER NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  position INTEGER NOT NULL CHECK (position > 0),
  part_id  INTEGER NOT NULL REFERENCES item(id) ON DELETE CASCADE,
  PRIMARY KEY (whole_id, position),
  CHECK (whole_id <> part_id)
);
CREATE INDEX part_of ON part (part_id);
