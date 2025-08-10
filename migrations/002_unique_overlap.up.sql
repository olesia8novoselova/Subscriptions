CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE subscriptions
  ADD COLUMN period daterange
  GENERATED ALWAYS AS (
    daterange(
      start_date,
      CASE
        WHEN end_date IS NULL THEN NULL
        ELSE (end_date + INTERVAL '1 month')::date
      END,
      '[)'
    )
  ) STORED;

-- Исключающее ограничение: для одного user_id и service_name запрещены пересекающиеся периоды
ALTER TABLE subscriptions
  ADD CONSTRAINT uniq_user_service_period
  EXCLUDE USING gist (
    user_id WITH =,
    (lower(service_name)) WITH =,
    period WITH &&
  );
