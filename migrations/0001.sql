BEGIN;
--run the following with superuser privileges:
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS vector;

-- migration table start

CREATE TABLE IF NOT EXISTS migrations (
        version INTEGER UNIQUE NOT NULL DEFAULT 1
);

INSERT INTO migrations (version) VALUES (1);

-- migration table end

-- chats table start

CREATE TABLE IF NOT EXISTS chats (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	version INTEGER NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

-- auto update modified_at
CREATE OR REPLACE FUNCTION chats_update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;

$$ language 'plpgsql';

CREATE TRIGGER chats_trigger_update_modified_at
BEFORE UPDATE ON chats
FOR EACH ROW
EXECUTE PROCEDURE chats_update_modified_at();

-- chats table end

-- messages table start

CREATE TABLE messages (
	id BIGSERIAL PRIMARY KEY,
	message_date timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	phone_number TEXT NOT NULL,
	message TEXT NOT NULL,
	attachment TEXT NOT NULL,
	enable_semantic_search BOOLEAN NOT NULL DEFAULT TRUE,
	chat_id BIGINT NOT NULL REFERENCES chats(id) DEFAULT 1,
	version INTEGER NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	UNIQUE(message_date, phone_number)
);

CREATE OR REPLACE FUNCTION messages_update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;

$$ language 'plpgsql';

CREATE TRIGGER messages_trigger_update_modified_at
BEFORE UPDATE ON messages
FOR EACH ROW
EXECUTE PROCEDURE messages_update_modified_at();
-- messages table end


-- phrases table start

CREATE TABLE IF NOT EXISTS phrases (
	id BIGSERIAL PRIMARY KEY,
	content TEXT NOT NULL,
	openai_embeddings vector(1536),
	st_embeddings vector(384),
	tokens INT NOT NULL DEFAULT 0,
	sequence INT NOT NULL DEFAULT 0,
	content_field TEXT NOT NULL,
	message_id BIGINT NOT NULL REFERENCES messages(id),
	version INTEGER NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE phrases
ADD CONSTRAINT chk_content_field
CHECK (content_field IN ('messages.message'));

-- auto update modified_at
CREATE OR REPLACE FUNCTION phrases_update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;

$$ language 'plpgsql';

CREATE TRIGGER phrases_trigger_update_modified_at
BEFORE UPDATE ON phrases
FOR EACH ROW
EXECUTE PROCEDURE phrases_update_modified_at();

-- phrases table end

COMMIT;
