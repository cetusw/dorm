ALTER TABLE `group`
    ADD COLUMN leader_id BINARY(16) DEFAULT NULL AFTER `group_id`,
    ADD CONSTRAINT fk_group_leader FOREIGN KEY (leader_id) REFERENCES user (user_id) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE dormitory
    ADD COLUMN leader_id BINARY(16) DEFAULT NULL AFTER `dormitory_id`,
    ADD CONSTRAINT fk_dormitory_leader FOREIGN KEY (leader_id) REFERENCES user (user_id) ON DELETE SET NULL ON UPDATE CASCADE;
