ALTER TABLE `task`
    ADD COLUMN `frequency` INT NOT NULL DEFAULT 7 AFTER `cost`;

UPDATE `task`
SET `frequency` = CASE `recurrence_interval`
    WHEN 0 THEN 0
    WHEN 1 THEN 7
    WHEN 2 THEN 14
    WHEN 4 THEN 28
    WHEN 12 THEN 84
    ELSE 7
END;

ALTER TABLE `task`
    DROP CONSTRAINT `chk_task_recurrence_interval_non_negative`;

ALTER TABLE `task`
    DROP COLUMN `start_sequence`,
    DROP COLUMN `recurrence_interval`;

ALTER TABLE `duty`
    DROP FOREIGN KEY `fk_duty_group`,
    DROP INDEX `uq_duty_group_sequence`,
    DROP COLUMN `sequence_number`,
    DROP COLUMN `group_id`;
