INSERT INTO area (area_id, area_floor, area_name, group_id)
VALUES (12, 1, '101 комната', UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', ''))),
       (13, 1, '103 комната', UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', '')));

INSERT INTO `task` (`task_id`, `area_id`, `task_title`, `task_cost`)
VALUES (UUID_TO_BIN(UUID()), 12, 'Пропылесосить полы', 2),
       (UUID_TO_BIN(UUID()), 12, 'Помыть полы', 3),
       (UUID_TO_BIN(UUID()), 12, 'Протереть полки от пыли', 1),
       (UUID_TO_BIN(UUID()), 12, 'Протереть тумбы от пыли', 1),
       (UUID_TO_BIN(UUID()), 12, 'Протереть столы от пыли', 1),
       (UUID_TO_BIN(UUID()), 13, 'Пропылесосить полы', 2),
       (UUID_TO_BIN(UUID()), 13, 'Помыть полы', 3),
       (UUID_TO_BIN(UUID()), 13, 'Протереть полки от пыли', 1),
       (UUID_TO_BIN(UUID()), 13, 'Протереть тумбы от пыли', 1),
       (UUID_TO_BIN(UUID()), 13, 'Протереть столы от пыли', 1);