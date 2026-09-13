-- Keep feature-flag tables isolated when using the Atlas PostgreSQL database.
CREATE SCHEMA IF NOT EXISTS feature_flags;
DO $$
BEGIN
  IF to_regclass('public.revision_counter') IS NOT NULL AND to_regclass('feature_flags.revision_counter') IS NULL THEN ALTER TABLE public.revision_counter SET SCHEMA feature_flags; END IF;
  IF to_regclass('public.feature_flags') IS NOT NULL AND to_regclass('feature_flags.feature_flags') IS NULL THEN ALTER TABLE public.feature_flags SET SCHEMA feature_flags; END IF;
  IF to_regclass('public.flag_history') IS NOT NULL AND to_regclass('feature_flags.flag_history') IS NULL THEN ALTER TABLE public.flag_history SET SCHEMA feature_flags; END IF;
  IF to_regclass('public.outbox_events') IS NOT NULL AND to_regclass('feature_flags.outbox_events') IS NULL THEN ALTER TABLE public.outbox_events SET SCHEMA feature_flags; END IF;
END $$;
SET search_path TO feature_flags, public;