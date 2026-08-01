ALTER TABLE `user`
    ADD COLUMN `login` VARCHAR(255) NULL AFTER `telegram_id`,
    ADD COLUMN `password_hash` VARCHAR(255) NULL AFTER `login`;

UPDATE `user`
SET
    `login` = COALESCE(NULLIF(`login`, ''), CONCAT('user_', LOWER(HEX(`id`)))),
    `password_hash` = COALESCE(NULLIF(`password_hash`, ''), 'pending_password_hash');

ALTER TABLE `user`
    MODIFY COLUMN `login` VARCHAR(255) NOT NULL,
    MODIFY COLUMN `password_hash` VARCHAR(255) NOT NULL;
