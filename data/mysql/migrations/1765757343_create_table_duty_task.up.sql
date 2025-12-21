CREATE TABLE `duty_task`
(
    `id`                BINARY(16) NOT NULL,
    `duty_id`           BINARY(16) NOT NULL,
    `task_id`           BINARY(16) NOT NULL,
    `assignee_id`       BINARY(16) DEFAULT NULL,
    `reviewer_id`       BINARY(16) DEFAULT NULL,
    `assignment_date`   TIMESTAMP  DEFAULT NULL,
    `completion_date`   TIMESTAMP  DEFAULT NULL,
    `verification_date` TIMESTAMP  DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_duty_task_per_duty` (`duty_id`, `task_id`),
    CONSTRAINT `fk_duty_task_duty` FOREIGN KEY (`duty_id`) REFERENCES `duty` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_duty_task_task` FOREIGN KEY (`task_id`) REFERENCES `task` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_duty_task_assignee` FOREIGN KEY (`assignee_id`) REFERENCES `user` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_duty_task_reviewer` FOREIGN KEY (`reviewer_id`) REFERENCES `user` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;