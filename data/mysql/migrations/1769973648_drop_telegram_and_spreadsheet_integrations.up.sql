ALTER TABLE `user`
    DROP INDEX `uq_telegram_id`,
    DROP COLUMN `telegram_id`;

ALTER TABLE `group`
    DROP COLUMN `spreadsheet_id`;
