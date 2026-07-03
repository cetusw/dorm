CREATE TABLE `duty_task_override`
(
    `task_id`              BINARY(16) NOT NULL,
    `include_in_next_duty` BOOLEAN    NOT NULL,
    PRIMARY KEY (`task_id`),
    CONSTRAINT `fk_duty_task_override_task` FOREIGN KEY (`task_id`) REFERENCES `task` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;
