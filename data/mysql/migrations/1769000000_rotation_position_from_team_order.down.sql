ALTER TABLE `group`
    ADD COLUMN next_duty_team INT DEFAULT NULL;

DROP INDEX idx_team_group_rotation ON team;

ALTER TABLE team
    DROP INDEX uq_team_group_rotation_position;

ALTER TABLE team
    CHANGE COLUMN rotation_position team_order INT NOT NULL;
