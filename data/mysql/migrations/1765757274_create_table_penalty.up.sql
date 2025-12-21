CREATE TABLE `penalty`
(
    `id`          BINARY(16) NOT NULL,
    `user_id`     BINARY(16) NOT NULL,
    `reason`      TEXT       NOT NULL,
    `resolution`  TEXT                DEFAULT NULL,
    `created_at`  TIMESTAMP  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `resolved_at` TIMESTAMP           DEFAULT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_penalty_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;
