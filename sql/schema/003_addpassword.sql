-- +goose Up
ALTER TABLE users 
add COLUMN password TEXT NOT NULL DEFAULT 'unset';

-- +goose Down
ALTER TABLE users
drop COLUMN password;