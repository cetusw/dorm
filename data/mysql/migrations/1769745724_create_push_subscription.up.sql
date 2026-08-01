CREATE TABLE `push_subscription`
(
    `id`            BINARY(16)   NOT NULL,
    `user_id`       BINARY(16)   NOT NULL,
    `endpoint`      TEXT         NOT NULL,
    `endpoint_hash` BINARY(32)   NOT NULL,
    `p256dh`        VARCHAR(255) NOT NULL,
    `auth_secret`   VARCHAR(255) NOT NULL,
    `user_agent`    VARCHAR(512) DEFAULT NULL,
    `created_at`    DATETIME     NOT NULL,
    `updated_at`    DATETIME     NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_push_subscription_endpoint_hash` (`endpoint_hash`),
    KEY `idx_push_subscription_user` (`user_id`),
    CONSTRAINT `fk_push_subscription_user`
        FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;
