CREATE TEMPORARY TABLE `team_leader_precondition` (
    `leader_id` BINARY(16) NOT NULL
);

INSERT INTO `team_leader_precondition` (`leader_id`)
SELECT `leader_id`
FROM `team`;

DROP TEMPORARY TABLE `team_leader_precondition`;

ALTER TABLE `team`
    DROP FOREIGN KEY `fk_team_leader`;

ALTER TABLE `team`
    MODIFY COLUMN `leader_id` BINARY(16) NOT NULL,
    DROP COLUMN `name`,
    ADD CONSTRAINT `fk_team_leader`
        FOREIGN KEY (`leader_id`) REFERENCES `user` (`id`)
        ON DELETE RESTRICT ON UPDATE CASCADE;
