ALTER TABLE outbox
ADD COLUMN idempotency_key VARCHAR(255);
