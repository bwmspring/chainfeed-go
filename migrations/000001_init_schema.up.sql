-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    wallet_address VARCHAR(42) NOT NULL UNIQUE,
    nonce VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_wallet_address ON users(wallet_address);

-- Watched addresses table
CREATE TABLE IF NOT EXISTS watched_addresses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address VARCHAR(42) NOT NULL,
    label VARCHAR(100),
    ens_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, address)
);

CREATE INDEX idx_watched_addresses_user_id ON watched_addresses(user_id);
CREATE INDEX idx_watched_addresses_address ON watched_addresses(address);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    tx_hash VARCHAR(66) NOT NULL UNIQUE,
    block_number BIGINT NOT NULL,
    block_timestamp TIMESTAMP NOT NULL,
    from_address VARCHAR(42) NOT NULL,
    to_address VARCHAR(42),
    value NUMERIC(78, 0) NOT NULL DEFAULT 0,
    tx_type VARCHAR(20) NOT NULL,
    token_address VARCHAR(42),
    token_id VARCHAR(78),
    token_symbol VARCHAR(20),
    token_decimals INT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_tx_hash ON transactions(tx_hash);
CREATE INDEX idx_transactions_from_address ON transactions(from_address);
CREATE INDEX idx_transactions_to_address ON transactions(to_address);
CREATE INDEX idx_transactions_block_timestamp ON transactions(block_timestamp DESC);

-- Feed items table
CREATE TABLE IF NOT EXISTS feed_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    transaction_id BIGINT NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    watched_address_id BIGINT NOT NULL REFERENCES watched_addresses(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, transaction_id)
);

CREATE INDEX idx_feed_items_user_id ON feed_items(user_id, created_at DESC);
CREATE INDEX idx_feed_items_transaction_id ON feed_items(transaction_id);
