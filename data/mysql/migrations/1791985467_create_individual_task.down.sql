DROP TABLE individual_task;

-- Rollback intentionally truncates reasons written by individual-task verification
-- that exceed the former 256-character schema limit.
UPDATE penalty_entry SET reason = LEFT(reason, 256) WHERE CHAR_LENGTH(reason) > 256;
ALTER TABLE penalty_entry
    MODIFY reason VARCHAR(256) NOT NULL;
