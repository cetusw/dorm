ALTER TABLE `task`
    ADD COLUMN `recurrence_interval` INT NOT NULL DEFAULT 1 AFTER `cost`,
    ADD COLUMN `start_sequence` INT NOT NULL DEFAULT 1 AFTER `recurrence_interval`,
    ADD CONSTRAINT `chk_task_recurrence_interval_non_negative` CHECK (`recurrence_interval` >= 0);

UPDATE `task`
SET `recurrence_interval` = CASE
    WHEN `frequency` IN (0, 1, 2, 4, 12) THEN `frequency`
    WHEN `frequency` = 7 THEN 1
    WHEN `frequency` = 14 THEN 2
    WHEN `frequency` IN (28, 30) THEN 4
    WHEN `frequency` IN (84, 90) THEN 12
    WHEN `frequency` < 0 THEN 1
    ELSE 1
END,
    `start_sequence` = 1;

ALTER TABLE `task`
    DROP COLUMN `frequency`;

ALTER TABLE `duty`
    ADD COLUMN `group_id` BINARY(16) NULL AFTER `team_id`,
    ADD COLUMN `sequence_number` INT NULL AFTER `end_date`;

UPDATE `duty` d
JOIN `team` t ON t.id = d.team_id
SET d.group_id = t.group_id;

UPDATE `duty` d
JOIN (
    SELECT ranked.id,
           ranked.group_id,
           ranked.sequence_number
    FROM (
        SELECT d2.id,
               t2.group_id,
               ROW_NUMBER() OVER (
                   PARTITION BY t2.group_id
                   ORDER BY d2.start_date ASC, d2.id ASC
               ) AS sequence_number
        FROM `duty` d2
        JOIN `team` t2 ON t2.id = d2.team_id
    ) AS ranked
) seq ON seq.id = d.id
SET d.sequence_number = seq.sequence_number,
    d.group_id = seq.group_id;

ALTER TABLE `duty`
    MODIFY COLUMN `group_id` BINARY(16) NOT NULL,
    MODIFY COLUMN `sequence_number` INT NOT NULL,
    ADD CONSTRAINT `fk_duty_group` FOREIGN KEY (`group_id`) REFERENCES `group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    ADD CONSTRAINT `uq_duty_group_sequence` UNIQUE KEY (`group_id`, `sequence_number`);
