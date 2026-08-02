ALTER TABLE `penalty`
    DROP COLUMN `resolution`,
    MODIFY COLUMN `reason` VARCHAR(256) NOT NULL,
    ADD COLUMN `weight` DECIMAL(10,1) NOT NULL AFTER `user_id`,
    ADD COLUMN `issued_on` DATE NOT NULL AFTER `reason`,
    ADD CONSTRAINT `chk_penalty_weight` CHECK (`weight` >= 0.1);

CREATE INDEX `idx_penalty_user_resolved`
    ON `penalty` (`user_id`, `resolved_at`);
