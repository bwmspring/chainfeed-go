-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    wallet_address VARCHAR(42) UNIQUE NOT NULL,
    nonce VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_wallet_address ON users(wallet_address);

-- Create watched_addresses table
CREATE TABLE IF NOT EXISTS watched_addresses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address VARCHAR(42) NOT NULL,
    label VARCHAR(255),
    ens_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, address)
);

CREATE INDEX IF NOT EXISTS idx_watched_addresses_user_id ON watched_addresses(user_id);
CREATE INDEX IF NOT EXISTS idx_watched_addresses_address ON watched_addresses(address);

-- Create feed_items table (for future use)
CREATE TABLE IF NOT EXISTS feed_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    transaction_id BIGINT NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    watched_address_id BIGINT NOT NULL REFERENCES watched_addresses(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feed_items_user_id ON feed_items(user_id);
CREATE INDEX IF NOT EXISTS idx_feed_items_created_at ON feed_items(created_at DESC);
