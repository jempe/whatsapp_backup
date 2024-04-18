BEGIN;

CREATE TABLE messages (
	id BIGSERIAL PRIMARY KEY,
	message_date timestamp(0) with time zone NOT NULL DEFAULT NOW(),
	phone_number TEXT NOT NULL,
	message TEXT NOT NULL,
	attachment TEXT NOT NULL,
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
--

COMMIT;
