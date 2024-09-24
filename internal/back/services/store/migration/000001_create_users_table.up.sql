-- user authorization
CREATE TABLE users (
    id SERIAL PRIMARY KEY,               	-- unique user identifier
    email VARCHAR(255) UNIQUE,  			-- email (login)
    password_hash TEXT,         			-- password hash (for storing encrypted password)
    created_at TIMESTAMPTZ DEFAULT NOW(),	-- account creation date
    last_login TIMESTAMPTZ,              	-- last login time
    token_expiration TIMESTAMPTZ         	-- JWT token expiration time
);

-- global task queue
CREATE TABLE queue (
    id SERIAL PRIMARY KEY,   			    -- unique queue item identifier
    order_num SERIAL              			-- order number (auto-increment)
);

-- user files
CREATE TABLE userfiles (
    id SERIAL PRIMARY KEY,            		-- unique file identifier
    queue_id INTEGER REFERENCES queue(id),  -- reference to queue (can be NULL)
    user_id INTEGER NOT NULL,          		-- user identifier
    order_num SERIAL,                  		-- order number
    src_file_url TEXT NOT NULL,             -- URL of the source file
    src_file_key TEXT NOT NULL,             -- key of the source file
    dest_file_url TEXT NOT NULL,            -- URL of the processed file
    dest_file_key TEXT NOT NULL,            -- key of the processed file
    state TEXT NOT NULL,				    -- task state: PENDING/PROCESSED/ERROR/OUTDATED
    created_at TIMESTAMPTZ DEFAULT NOW(),   -- record creation time

    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_userfiles_user_id_order_num ON userfiles(user_id, order_num);
CREATE INDEX idx_userfiles_queue_id ON userfiles(queue_id);

-- transaction outbox for sending to the message queue
CREATE TABLE outbox (
    id SERIAL PRIMARY KEY,                  -- unique message identifier
    payload JSONB NOT NULL,                 -- data for transmission (data for Kafka)
    status VARCHAR(50) DEFAULT 'PENDING',   -- sending status (PENDING, SENT, FAILED)
    created_at TIMESTAMPTZ DEFAULT NOW(),   -- record creation time
    processed_at TIMESTAMPTZ,               -- processing time (when the message was sent)
);
