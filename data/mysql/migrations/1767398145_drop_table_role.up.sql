ALTER TABLE `user` DROP FOREIGN KEY `fk_user_role`;
ALTER TABLE `user` DROP COLUMN `role_id`;
DROP TABLE `role`;