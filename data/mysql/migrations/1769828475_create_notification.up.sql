CREATE TABLE `notification`
(
    `id`                BINARY(16)    NOT NULL,
    `user_id`           BINARY(16)    NOT NULL,
    `type`              VARCHAR(64)   NOT NULL,
    `title`             VARCHAR(255)  NOT NULL,
    `body`              TEXT          NOT NULL,
    `target_url`        VARCHAR(1024) NOT NULL,
    `deduplication_key` VARCHAR(255)  NOT NULL,
    `created_at`        DATETIME      NOT NULL,
    `read_at`           DATETIME      DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_notification_user_deduplication` (`user_id`, `deduplication_key`),
    KEY `idx_notification_user_created` (`user_id`, `created_at`),
    CONSTRAINT `fk_notification_user`
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;
