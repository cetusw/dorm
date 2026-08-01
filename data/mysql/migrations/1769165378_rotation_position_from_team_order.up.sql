SET @needs_normalization := (
    SELECT CASE
        WHEN EXISTS (
            SELECT 1
            FROM (
                SELECT group_id
                FROM team
                GROUP BY group_id, team_order
                HAVING COUNT(*) > 1
                    OR team_order IS NULL
                    OR team_order <= 0
            ) duplicated_positions
        ) THEN 1
        ELSE 0
    END
);

SET @normalize_sql := IF(
    @needs_normalization = 1,
    'UPDATE team t
     JOIN (
         SELECT id, ROW_NUMBER() OVER (PARTITION BY group_id ORDER BY COALESCE(team_order, 2147483647), id) AS normalized_position
         FROM team
     ) normalized ON normalized.id = t.id
     SET t.team_order = normalized.normalized_position',
    'DO 0'
);
PREPARE normalize_stmt FROM @normalize_sql;
EXECUTE normalize_stmt;
DEALLOCATE PREPARE normalize_stmt;

ALTER TABLE team
    CHANGE COLUMN team_order rotation_position INT NOT NULL;

ALTER TABLE team
    ADD CONSTRAINT uq_team_group_rotation_position
        UNIQUE (group_id, rotation_position);

CREATE INDEX idx_team_group_rotation
    ON team (group_id, rotation_position);

ALTER TABLE `group`
    DROP COLUMN next_duty_team;
