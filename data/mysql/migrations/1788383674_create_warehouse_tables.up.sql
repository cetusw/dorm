CREATE TABLE `warehouse_item`
(
    `id`           BINARY(16)   NOT NULL,
    `dormitory_id` INT UNSIGNED NOT NULL,
    `name`         VARCHAR(255) NOT NULL,
    `created_at`   DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_warehouse_item_dormitory_name` (`dormitory_id`, `name`),
    CONSTRAINT `fk_warehouse_item_dormitory`
        FOREIGN KEY (`dormitory_id`) REFERENCES `dormitory` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;

CREATE TABLE `warehouse_movement`
(
    `id`         BINARY(16)                  NOT NULL,
    `item_id`    BINARY(16)                  NOT NULL,
    `type`       ENUM ('add', 'write-off')   NOT NULL,
    `quantity`   INT UNSIGNED                NOT NULL,
    `comment`    VARCHAR(256)                DEFAULT NULL,
    `created_by` BINARY(16)                  NOT NULL,
    `created_at` DATETIME(6)                 NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (`id`),
    KEY `idx_warehouse_movement_item_created` (`item_id`, `created_at`, `id`),
    CONSTRAINT `chk_warehouse_movement_quantity`
        CHECK (`quantity` > 0),
    CONSTRAINT `fk_warehouse_movement_item`
        FOREIGN KEY (`item_id`) REFERENCES `warehouse_item` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_warehouse_movement_created_by`
        FOREIGN KEY (`created_by`) REFERENCES `user` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;
