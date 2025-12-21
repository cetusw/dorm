CREATE TABLE `user`
(
    `id`           BINARY(16)   NOT NULL,
    `telegram_id`  BIGINT       NOT NULL,
    `first_name`   VARCHAR(255) NOT NULL,
    `last_name`    VARCHAR(255) NOT NULL,
    `middle_name`  VARCHAR(255) DEFAULT NULL,
    `team_id`      BINARY(16)   DEFAULT NULL,
    `floor_number` INT          DEFAULT NULL,
    `room_number`  VARCHAR(50)  DEFAULT NULL,
    `dormitory_id` INT UNSIGNED DEFAULT NULL,
    `role_id`      INT UNSIGNED NOT NULL,
    `created_at`   TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    `deleted_at`   TIMESTAMP    DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_telegram_id` (`telegram_id`),
    CONSTRAINT `fk_user_team` FOREIGN KEY (`team_id`) REFERENCES `team` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT `fk_user_dormitory` FOREIGN KEY (`dormitory_id`) REFERENCES `dormitory` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_user_role` FOREIGN KEY (`role_id`) REFERENCES `role` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;
