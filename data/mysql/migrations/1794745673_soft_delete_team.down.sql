CREATE TEMPORARY TABLE `team_soft_delete_rollback_precondition` (
    `rotation_position` INT NOT NULL
);

INSERT INTO `team_soft_delete_rollback_precondition` (`rotation_position`)
SELECT `rotation_position`
FROM `team`
WHERE `deleted_at` IS NOT NULL;

DROP TEMPORARY TABLE `team_soft_delete_rollback_precondition`;

ALTER TABLE `duty`
    DROP FOREIGN KEY `fk_duty_team`;

ALTER TABLE `duty`
    ADD CONSTRAINT `fk_duty_team`
        FOREIGN KEY (`team_id`) REFERENCES `team` (`id`)
        ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `team`
    MODIFY COLUMN `rotation_position` INT NOT NULL,
    DROP COLUMN `deleted_at`;
