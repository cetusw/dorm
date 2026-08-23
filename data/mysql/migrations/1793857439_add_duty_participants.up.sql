ALTER TABLE `duty`
    ADD COLUMN `leader_id` BINARY(16) NULL AFTER `team_id`,
    ADD CONSTRAINT `fk_duty_leader` FOREIGN KEY (`leader_id`) REFERENCES `user` (`id`) ON DELETE SET NULL ON UPDATE CASCADE;

UPDATE `duty` d
JOIN `team` t ON t.id = d.team_id
SET d.leader_id = t.leader_id;

CREATE TABLE `duty_participants`
(
    `duty_id`        BINARY(16) NOT NULL,
    `participant_id` BINARY(16) NOT NULL,
    `type`           VARCHAR(16) NOT NULL,
    `excluded_at`    TIMESTAMP NULL,
    PRIMARY KEY (`duty_id`, `participant_id`),
    KEY `idx_duty_participants_participant` (`participant_id`),
    CONSTRAINT `fk_duty_participants_duty` FOREIGN KEY (`duty_id`) REFERENCES `duty` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_duty_participants_user` FOREIGN KEY (`participant_id`) REFERENCES `user` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT `chk_duty_participant_type` CHECK (`type` IN ('REGULAR', 'TEMPORARY'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO `duty_participants` (`duty_id`, `participant_id`, `type`, `excluded_at`)
SELECT d.id, u.id, 'REGULAR', NULL
FROM duty d
JOIN user u ON u.team_id = d.team_id AND u.deleted_at IS NULL
WHERE d.end_date > CURRENT_TIMESTAMP
   OR EXISTS (SELECT 1 FROM duty_task dt WHERE dt.duty_id = d.id AND dt.verification_date IS NULL)
UNION
SELECT d.id, d.leader_id, 'REGULAR', NULL
FROM duty d
WHERE d.leader_id IS NOT NULL
  AND (d.end_date > CURRENT_TIMESTAMP
       OR EXISTS (SELECT 1 FROM duty_task dt WHERE dt.duty_id = d.id AND dt.verification_date IS NULL));
