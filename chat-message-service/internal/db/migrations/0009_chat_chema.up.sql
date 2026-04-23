BEGIN;

-- three steps to allow migrations to proceed

-- 1. Add the column (Allowing NULLs for a split second)
ALTER TABLE conversations ADD COLUMN created_by UUID;

-- 2. Fill the existing rows with the "Nil" UUID
UPDATE conversations SET created_by = '00000000-0000-0000-0000-000000000000';

-- 3. Enforce the NOT NULL constraint (Now that rows are filled)
ALTER TABLE conversations ALTER COLUMN created_by SET NOT NULL;

COMMIT;
