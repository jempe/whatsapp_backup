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

-- contacts table start

CREATE TABLE IF NOT EXISTS contacts (
	id BIGSERIAL PRIMARY KEY,
	phone_number TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	version INTEGER NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

-- auto update modified_at

CREATE OR REPLACE FUNCTION contacts_update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;

$$ language 'plpgsql';

CREATE TRIGGER contacts_trigger_update_modified_at
BEFORE UPDATE ON contacts
FOR EACH ROW
EXECUTE PROCEDURE contacts_update_modified_at();

-- contacts table end

INSERT INTO contacts (id, phone_number, name) VALUES (0, '0', 'Group');

-- chats table start

CREATE TABLE IF NOT EXISTS chats (
	id BIGSERIAL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	contact_id BIGINT NOT NULL REFERENCES contacts(id),
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
	contact_id BIGINT NOT NULL REFERENCES contacts(id),
	message TEXT NOT NULL,
	attachment TEXT NOT NULL,
	enable_semantic_search BOOLEAN NOT NULL DEFAULT TRUE,
	chat_id BIGINT NOT NULL REFERENCES chats(id) DEFAULT 1,
	version INTEGER NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	UNIQUE(message_date, contact_id)
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

-- user table start

CREATE TABLE IF NOT EXISTS users (
	id bigserial PRIMARY KEY,
	name text NOT NULL,
	email citext UNIQUE NOT NULL,
	password_hash bytea NOT NULL,
	activated bool NOT NULL,
	version integer NOT NULL DEFAULT 1,
	created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	modified_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

-- auto update modified_at
CREATE OR REPLACE FUNCTION users_update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;

$$ language 'plpgsql';

CREATE TRIGGER users_trigger_update_modified_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE PROCEDURE users_update_modified_at();

-- user table end

-- tokens table start

CREATE TABLE IF NOT EXISTS tokens (
	hash bytea PRIMARY KEY,
	user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
	expiry timestamp(0) with time zone NOT NULL,
	scope text NOT NULL
);

-- tokens table end

COMMIT;
