CREATE TABLE `area`
(
    `id`       INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `floor`    INT        DEFAULT NULL,
    `name`     VARCHAR(255) NOT NULL,
    `group_id` BINARY(16) DEFAULT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_area_group` FOREIGN KEY (`group_id`) REFERENCES `group` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;
