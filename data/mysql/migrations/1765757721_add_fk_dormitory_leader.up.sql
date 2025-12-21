ALTER TABLE `dormitory`
    ADD CONSTRAINT fk_dormitory_leader FOREIGN KEY (leader_id) REFERENCES user (id) ON DELETE SET NULL ON UPDATE CASCADE;
