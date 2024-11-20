-- +goose Up
CREATE TABLE chirps (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  body TEXT NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

ALTER TABLE users 
ALTER COLUMN id SET DEFAULT gen_random_uuid();
;

-- +goose Down
DROP TABLE chirps;

ALTER TABLE users
ALTER COLUMN id DROP DEFAULT;