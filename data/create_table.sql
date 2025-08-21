SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS `area`;
CREATE TABLE `area`
(
    `area_id`    INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `area_floor` INT DEFAULT NULL,
    `area_name`  VARCHAR(255) NOT NULL,
    PRIMARY KEY (`area_id`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;

DROP TABLE IF EXISTS `dormitory`;
CREATE TABLE `dormitory`
(
    `dormitory_id`           INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `dormitory_name`         VARCHAR(255) NOT NULL,
    `dormitory_city`         VARCHAR(255) NOT NULL,
    `dormitory_street`       VARCHAR(255) NOT NULL,
    `dormitory_house_number` VARCHAR(50)  NOT NULL,
    PRIMARY KEY (`dormitory_id`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;

DROP TABLE IF EXISTS `role`;
CREATE TABLE `role`
(
    `role_id`          INT UNSIGNED NOT NULL,
    `role_name`        VARCHAR(255) NOT NULL,
    `role_description` TEXT DEFAULT NULL,
    PRIMARY KEY (`role_id`),
    UNIQUE KEY `uq_role_name` (`role_name`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;

DROP TABLE IF EXISTS `team`;
CREATE TABLE `team`
(
    `team_id`        INT UNSIGNED NOT NULL,
    `team_leader_id` BINARY(16) DEFAULT NULL,
    `team_color`     VARCHAR(6) DEFAULT NULL,
    PRIMARY KEY (`team_id`)
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;

DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`
(
    `user_id`      BINARY(16)   NOT NULL,
    `telegram_id`  BIGINT       NOT NULL,
    `first_name`   VARCHAR(255) NOT NULL,
    `last_name`    VARCHAR(255) NOT NULL,
    `middle_name`  VARCHAR(255) DEFAULT NULL,
    `team_id`      INT UNSIGNED DEFAULT NULL,
    `room_number`  VARCHAR(50)  DEFAULT NULL,
    `dormitory_id` INT UNSIGNED DEFAULT NULL,
    `role_id`      INT UNSIGNED NOT NULL,
    `created_at`   TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    `deleted_at`   TIMESTAMP    DEFAULT NULL,
    PRIMARY KEY (`user_id`),
    UNIQUE KEY `uq_telegram_id` (`telegram_id`),
    CONSTRAINT `fk_user_team` FOREIGN KEY (`team_id`) REFERENCES `team` (`team_id`) ON DELETE SET NULL ON UPDATE CASCADE,
    CONSTRAINT `fk_user_dormitory` FOREIGN KEY (`dormitory_id`) REFERENCES `dormitory` (`dormitory_id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_user_role` FOREIGN KEY (`role_id`) REFERENCES `role` (`role_id`) ON DELETE RESTRICT ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;

ALTER TABLE `team`
    ADD CONSTRAINT `fk_team_leader` FOREIGN KEY (`team_leader_id`) REFERENCES `user` (`user_id`) ON DELETE SET NULL ON UPDATE CASCADE;

DROP TABLE IF EXISTS `task`;
CREATE TABLE `task`
(
    `task_id`    BINARY(16)   NOT NULL,
    `area_id`    INT UNSIGNED NOT NULL,
    `task_title` VARCHAR(255) NOT NULL,
    `task_cost`  INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (`task_id`),
    CONSTRAINT `fk_task_area` FOREIGN KEY (`area_id`) REFERENCES `area` (`area_id`) ON DELETE RESTRICT ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;

DROP TABLE IF EXISTS `duty`;
CREATE TABLE `duty`
(
    `duty_id`         BINARY(16)   NOT NULL,
    `team_id`         INT UNSIGNED NOT NULL,
    `duty_start_date` TIMESTAMP    NOT NULL,
    `duty_end_date`   TIMESTAMP    NOT NULL,
    PRIMARY KEY (`duty_id`),
    CONSTRAINT `fk_duty_team` FOREIGN KEY (`team_id`) REFERENCES `team` (`team_id`) ON DELETE CASCADE ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;

DROP TABLE IF EXISTS `penalty`;
CREATE TABLE `penalty`
(
    `penalty_id`         BINARY(16) NOT NULL,
    `user_id`            BINARY(16) NOT NULL,
    `penalty_reason`     TEXT       NOT NULL,
    `penalty_resolution` TEXT                DEFAULT NULL,
    `created_at`         TIMESTAMP  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `resolved_at`        TIMESTAMP           DEFAULT NULL,
    PRIMARY KEY (`penalty_id`),
    CONSTRAINT `fk_penalty_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;

DROP TABLE IF EXISTS `user_availability`;
CREATE TABLE `user_availability`
(
    `availability_id`   BINARY(16) NOT NULL,
    `user_id`           BINARY(16) NOT NULL,
    `start_date`        TIMESTAMP  NOT NULL,
    `end_date`          TIMESTAMP  NOT NULL,
    `is_user_available` BOOLEAN    NOT NULL DEFAULT TRUE,
    PRIMARY KEY (`availability_id`),
    CONSTRAINT `fk_availability_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;

DROP TABLE IF EXISTS `duty_task`;
CREATE TABLE `duty_task`
(
    `duty_task_id`      BINARY(16) NOT NULL,
    `duty_id`           BINARY(16) NOT NULL,
    `task_id`           BINARY(16) NOT NULL,
    `assignee_id`       BINARY(16) DEFAULT NULL,
    `reviewer_id`       BINARY(16) DEFAULT NULL,
    `assignment_date`   TIMESTAMP  DEFAULT NULL,
    `completion_date`   TIMESTAMP  DEFAULT NULL,
    `verification_date` TIMESTAMP  DEFAULT NULL,
    PRIMARY KEY (`duty_task_id`),
    UNIQUE KEY `uq_duty_task_per_duty` (`duty_id`, `task_id`),
    CONSTRAINT `fk_duty_task_duty` FOREIGN KEY (`duty_id`) REFERENCES `duty` (`duty_id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_duty_task_task` FOREIGN KEY (`task_id`) REFERENCES `task` (`task_id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_duty_task_assignee` FOREIGN KEY (`assignee_id`) REFERENCES `user` (`user_id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `fk_duty_task_reviewer` FOREIGN KEY (`reviewer_id`) REFERENCES `user` (`user_id`) ON DELETE SET NULL ON UPDATE CASCADE
)
    ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_unicode_ci
;

SET FOREIGN_KEY_CHECKS = 1;