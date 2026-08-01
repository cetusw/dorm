SET @manual_telegram_id := 0;

UPDATE `user`
SET `telegram_id` = (@manual_telegram_id := @manual_telegram_id - 1)
WHERE `telegram_id` IS NULL;

ALTER TABLE `user` MODIFY COLUMN `telegram_id` BIGINT NOT NULL;
