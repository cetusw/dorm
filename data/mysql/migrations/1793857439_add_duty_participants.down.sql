DROP TABLE `duty_participants`;
ALTER TABLE `duty` DROP FOREIGN KEY `fk_duty_leader`;
ALTER TABLE `duty` DROP COLUMN `leader_id`;
