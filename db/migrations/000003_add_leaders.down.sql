ALTER TABLE `group`
    DROP FOREIGN KEY fk_group_leader,
    DROP COLUMN leader_id;
ALTER TABLE dormitory
    DROP FOREIGN KEY fk_dormitory_leader,
    DROP COLUMN leader_id;
