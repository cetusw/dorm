CREATE INDEX `idx_duty_group_period`
    ON `duty` (`group_id`, `start_date`, `end_date`);

CREATE INDEX `idx_duty_team_sequence`
    ON `duty` (`team_id`, `sequence_number`);
