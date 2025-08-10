ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS uniq_user_service_period;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS period;
