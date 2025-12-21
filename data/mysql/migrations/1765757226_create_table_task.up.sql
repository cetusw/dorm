CREATE TABLE `task`
(
    `id`        BINARY(16)   NOT NULL,
    `area_id`   INT UNSIGNED NOT NULL,
    `title`     VARCHAR(255) NOT NULL,
    `cost`      INT          NOT NULL DEFAULT 0,
    `frequency` INT          NOT NULL DEFAULT 7,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_task_area` FOREIGN KEY (`area_id`) REFERENCES `area` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;
