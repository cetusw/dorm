CREATE TABLE `penalty_entry`
(
    `id`         BINARY(16)                         NOT NULL,
    `user_id`    BINARY(16)                         NOT NULL,
    `type`       ENUM ('ISSUE', 'RESOLVE')         NOT NULL,
    `weight`     DECIMAL(10, 1)                    NOT NULL,
    `reason`     VARCHAR(256)                      NOT NULL,
    `created_at` TIMESTAMP                         NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_penalty_entry_user_created` (`user_id`, `created_at`, `id`),
    CONSTRAINT `fk_penalty_entry_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `chk_penalty_entry_weight` CHECK (`weight` >= 0.1)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;

INSERT INTO `penalty_entry` (`id`, `user_id`, `type`, `weight`, `reason`, `created_at`)
SELECT
    `id`,
    `user_id`,
    'ISSUE',
    `weight`,
    `reason`,
    `created_at`
FROM `penalty`;

INSERT INTO `penalty_entry` (`id`, `user_id`, `type`, `weight`, `reason`, `created_at`)
SELECT
    UNHEX(REPLACE(UUID(), '-', '')),
    `user_id`,
    'RESOLVE',
    `weight`,
    'Предупреждение было снято',
    `resolved_at`
FROM `penalty`
WHERE `resolved_at` IS NOT NULL;

DROP TABLE `penalty`;
