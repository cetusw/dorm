ALTER TABLE `team`
    DROP FOREIGN KEY `fk_team_leader`,
    ADD COLUMN `name` VARCHAR(255) NULL AFTER `id`;

UPDATE `team` t
JOIN `user` u ON u.id = t.leader_id
SET t.name = TRIM(CONCAT(u.last_name, ' ', u.first_name));

ALTER TABLE `team`
    MODIFY COLUMN `name` VARCHAR(255) NOT NULL,
    MODIFY COLUMN `leader_id` BINARY(16) NULL,
    ADD CONSTRAINT `fk_team_leader`
        FOREIGN KEY (`leader_id`) REFERENCES `user` (`id`)
        ON DELETE SET NULL ON UPDATE CASCADE;
