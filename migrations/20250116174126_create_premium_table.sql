-- +goose Up
-- +goose StatementBegin
CREATE TYPE package_type AS ENUM ('remove_quota', 'verified_label');

CREATE TABLE premium_packages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),  -- Unique ID for each premium package
    user_id UUID NOT NULL,                            -- User ID as a foreign key
    package_type package_type NOT NULL,               -- The type of the premium package
    active_until DATE NOT NULL,                       -- The date until the package is active
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,   -- Optional: Timestamp for record creation
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,   -- Optional: Timestamp for record updates
    CONSTRAINT fk_user FOREIGN KEY (user_id)         -- Foreign key constraint for user_id
        REFERENCES users(id) ON DELETE CASCADE       -- Cascades delete when user is deleted
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop the foreign key constraint first, if any
ALTER TABLE premium_packages
    DROP CONSTRAINT IF EXISTS fk_user;

-- Drop the `premium_packages` table
DROP TABLE IF EXISTS premium_packages;

-- Drop the `package_type` enum type if it's not used elsewhere
DROP TYPE IF EXISTS package_type;
-- +goose StatementEnd
