-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION insert_default_profile()
RETURNS TRIGGER AS $$
BEGIN
    -- Insert a default profile for the new user
    INSERT INTO profiles (username, user_id)
    VALUES (NEW.username, NEW.id);
    
    -- Return the new row (this is required by the trigger function)
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger that fires after an insert on the users table
CREATE TRIGGER trigger_insert_default_profile
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION insert_default_profile();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop the trigger that fires after an insert on the users table
DROP TRIGGER IF EXISTS trigger_insert_default_profile ON users;

-- Drop the trigger function
DROP FUNCTION IF EXISTS insert_default_profile;
-- +goose StatementEnd
