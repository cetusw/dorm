ALTER TABLE `user`
    ADD COLUMN `telegram_id` BIGINT NULL AFTER `id`,
    ADD UNIQUE KEY `uq_telegram_id` (`telegram_id`);

ALTER TABLE `group`
    ADD COLUMN `spreadsheet_id` VARCHAR(255) NULL;
