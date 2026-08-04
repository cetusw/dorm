DROP INDEX `idx_penalty_user_resolved` ON `penalty`;

ALTER TABLE `penalty`
    DROP CHECK `chk_penalty_weight`,
    DROP COLUMN `weight`,
    DROP COLUMN `issued_on`,
    MODIFY COLUMN `reason` TEXT NOT NULL,
    ADD COLUMN `resolution` TEXT NULL AFTER `reason`;
