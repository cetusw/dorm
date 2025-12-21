CREATE TABLE `duty`
(
    `id`         BINARY(16) NOT NULL,
    `team_id`    BINARY(16) NOT NULL,
    `start_date` TIMESTAMP  NOT NULL,
    `end_date`   TIMESTAMP  NOT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_duty_team` FOREIGN KEY (`team_id`) REFERENCES `team` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;
