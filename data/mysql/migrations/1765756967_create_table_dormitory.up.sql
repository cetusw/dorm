CREATE TABLE `dormitory`
(
    `id`           INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `leader_id`    BINARY(16) DEFAULT NULL,
    `name`         VARCHAR(255) NOT NULL,
    `city`         VARCHAR(255) NOT NULL,
    `street_type`  VARCHAR(100) NOT NULL,
    `street_name`  VARCHAR(100) NOT NULL,
    `house_number` VARCHAR(50)  NOT NULL,
    PRIMARY KEY (`id`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;
