CREATE TABLE `team`
(
    `id`         BINARY(16) NOT NULL,
    `group_id`   BINARY(16) NOT NULL,
    `leader_id`  BINARY(16)   DEFAULT NULL,
    `color`      VARCHAR(6)   DEFAULT NULL,
    `team_order` INT UNSIGNED DEFAULT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_team_group` FOREIGN KEY (`group_id`) REFERENCES `group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;
