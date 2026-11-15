ALTER TABLE entry ADD COLUMN time_spent TEXT CHECK (time_spent IN ('not-enough', 'a-bit', 'just-right', 'a-lot', 'too-much'));
