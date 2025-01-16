-- +goose Up
-- +goose StatementBegin
-- Add the user_id column to the profiles table, referencing the users table
ALTER TABLE profiles
ADD COLUMN user_id UUID;

-- Add foreign key constraint on user_id referencing users(id)
ALTER TABLE profiles
ADD CONSTRAINT fk_profiles_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;


ALTER TABLE swipes ADD COLUMN created_at_date DATE GENERATED ALWAYS AS (DATE(created_at)) STORED;
CREATE INDEX idx_user_id_created_at ON swipes (user_id, created_at);
CREATE INDEX idx_user_id_created_at_date ON swipes (user_id, created_at_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_id_created_at_date;
ALTER TABLE swipes DROP COLUMN IF EXISTS created_at_date;
DROP INDEX IF EXISTS idx_user_id_created_at;
-- +goose StatementEnd
