CREATE TABLE `user_availability`
(
    `availability_id` BINARY(16) NOT NULL,
    `user_id`         BINARY(16) NOT NULL,
    `start_date`      TIMESTAMP  NOT NULL,
    `end_date`        TIMESTAMP  NOT NULL,
    `is_available`    BOOLEAN    NOT NULL DEFAULT TRUE,
    PRIMARY KEY (`availability_id`),
    CONSTRAINT `fk_availability_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;
