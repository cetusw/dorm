CREATE TABLE `penalty`
(
    `id`          BINARY(16)   NOT NULL,
    `user_id`     BINARY(16)   NOT NULL,
    `weight`      DECIMAL(10, 1) NOT NULL,
    `reason`      VARCHAR(256) NOT NULL,
    `issued_on`   DATE         NOT NULL,
    `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `resolved_at` TIMESTAMP             DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_penalty_user_resolved` (`user_id`, `resolved_at`),
    CONSTRAINT `fk_penalty_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `chk_penalty_weight` CHECK (`weight` >= 0.1)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;

INSERT INTO `penalty` (`id`, `user_id`, `weight`, `reason`, `issued_on`, `created_at`, `resolved_at`)
SELECT
    `id`,
    `user_id`,
    `weight`,
    `reason`,
    DATE(`created_at`),
    `created_at`,
    NULL
FROM `penalty_entry`
WHERE `type` = 'ISSUE';

DROP TABLE `penalty_entry`;
