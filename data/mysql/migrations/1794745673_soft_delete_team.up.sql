SET @team_deleted_at_exists := (
    SELECT COUNT(*)
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'team'
      AND COLUMN_NAME = 'deleted_at'
);
SET @team_deleted_at_sql := IF(
    @team_deleted_at_exists = 0,
    'ALTER TABLE `team` ADD COLUMN `deleted_at` TIMESTAMP NULL DEFAULT NULL AFTER `rotation_position`',
    'SELECT 1'
);
PREPARE team_deleted_at_statement FROM @team_deleted_at_sql;
EXECUTE team_deleted_at_statement;
DEALLOCATE PREPARE team_deleted_at_statement;

ALTER TABLE `team`
    MODIFY COLUMN `rotation_position` INT NULL;

ALTER TABLE `duty`
    DROP FOREIGN KEY `fk_duty_team`;

ALTER TABLE `duty`
    ADD CONSTRAINT `fk_duty_team`
        FOREIGN KEY (`team_id`) REFERENCES `team` (`id`)
        ON DELETE RESTRICT ON UPDATE CASCADE;
