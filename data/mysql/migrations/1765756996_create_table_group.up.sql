CREATE TABLE `group`
(
    `id`             BINARY(16)   NOT NULL,
    `leader_id`      BINARY(16)   DEFAULT NULL,
    `name`           VARCHAR(255) NOT NULL,
    `dormitory_id`   INT UNSIGNED NOT NULL,
    `spreadsheet_id` VARCHAR(255) DEFAULT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_team_group_dormitory` FOREIGN KEY (`dormitory_id`) REFERENCES `dormitory` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;
