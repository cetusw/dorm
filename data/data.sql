INSERT INTO role (role_id, role_name, role_description)
VALUES (1,
        'Житель',
        'Живёт в коливинге');

INSERT INTO user (user_id, telegram_id, first_name, last_name, role_id)
VALUES (UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', '')),
        523562011,
        'Михаил',
        'Кугелев',
        1);

INSERT INTO user (user_id, telegram_id, first_name, last_name, role_id)
VALUES (UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', '')),
        523562012,
        'Тест',
        'Тестович',
        1);

INSERT INTO team (team_id, team_leader_id, team_color)
VALUES (1,
        UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', '')),
        'b6d7a8');

INSERT INTO team (team_id, team_leader_id, team_color)
VALUES (2,
        UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', '')),
        'b4a7d6');

UPDATE user
SET team_id = 1
WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', ''));
UPDATE user
SET team_id = 2
WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', ''));

INSERT INTO area (area_floor, area_name)
VALUES (3, 'Коридор');
INSERT INTO area (area_floor, area_name)
VALUES (3, 'КУИ');
INSERT INTO area (area_floor, area_name)
VALUES (1, 'Северная часть кухни');
INSERT INTO area (area_floor, area_name)
VALUES (1, 'Северный коридор');
INSERT INTO area (area_floor, area_name)
VALUES (1, 'КУИ');
INSERT INTO area (area_floor, area_name)
VALUES (1, 'Прихожая');
INSERT INTO area (area_floor, area_name)
VALUES (-1, 'Прачка');

INSERT INTO `task` (`task_id`, `area_id`, `task_title`, `task_cost`)
VALUES (UUID_TO_BIN(UUID()), 1, 'Пропылесосить полы', 2),
       (UUID_TO_BIN(UUID()), 1, 'Помыть полы', 8),
       (UUID_TO_BIN(UUID()), 1, 'Протереть плинтуса', 3),
       (UUID_TO_BIN(UUID()), 1, 'Разобрать старые вещи на сушилках', 1),
       (UUID_TO_BIN(UUID()), 1, 'Протереть батареи', 1),
       (UUID_TO_BIN(UUID()), 1, 'Протереть подоконники', 2),
       (UUID_TO_BIN(UUID()), 1, 'Протереть окна', 4),
       (UUID_TO_BIN(UUID()), 1, 'Протереть двери', 2),
       (UUID_TO_BIN(UUID()), 1, 'Протереть огнетушители', 1),

       (UUID_TO_BIN(UUID()), 2, 'Пропылесосить полы', 1),
       (UUID_TO_BIN(UUID()), 2, 'Помыть полы', 1),
       (UUID_TO_BIN(UUID()), 2, 'Протереть подоконник', 1),
       (UUID_TO_BIN(UUID()), 2, 'Помыть раковину', 1),

       (UUID_TO_BIN(UUID()), 4, 'Пропылесосить полы', 2),
       (UUID_TO_BIN(UUID()), 4, 'Помыть полы', 6),
       (UUID_TO_BIN(UUID()), 4, 'Протереть плинтуса', 3),
       (UUID_TO_BIN(UUID()), 4, 'Протереть батареи', 2),
       (UUID_TO_BIN(UUID()), 4, 'Протереть розетки и выключатели', 1),
       (UUID_TO_BIN(UUID()), 4, 'Помыть северную лестницу', 9),
       (UUID_TO_BIN(UUID()), 4, 'Протереть огнетушители', 1), -- У задачи не было баллов, я поставил 1

       (UUID_TO_BIN(UUID()), 3, 'Пропылесосить полы', 2),
       (UUID_TO_BIN(UUID()), 3, 'Помыть полы', 6),
       (UUID_TO_BIN(UUID()), 3, 'Протереть батареи', 2),
       (UUID_TO_BIN(UUID()), 3, 'Протереть подоконники', 2),
       (UUID_TO_BIN(UUID()), 3, 'Протереть розетки и выключатели', 1),
       (UUID_TO_BIN(UUID()), 3, 'Убрать мусор из раковины и помыть её', 3),
       (UUID_TO_BIN(UUID()), 3, 'Собрать и вынести мусор + помыть место под мусоркой', 2),
       (UUID_TO_BIN(UUID()), 3, 'Протереть столы', 1),
       (UUID_TO_BIN(UUID()), 3, 'Помыть кухонный гарнитур', 3),
       (UUID_TO_BIN(UUID()), 3, 'Разобрать посуду', 1),
       (UUID_TO_BIN(UUID()), 3, 'Помыть ёмкости для столовых приборов', 1),
       (UUID_TO_BIN(UUID()), 3, 'Протереть полки в шкафчиках(где грязно)', 1),
       (UUID_TO_BIN(UUID()), 3, 'Постирать тряпки', 1),
       (UUID_TO_BIN(UUID()), 3, 'Помыть холодильники снаружи', 3),
       (UUID_TO_BIN(UUID()), 3, 'Избавиться от просрочки в холодильниках', 3),
       (UUID_TO_BIN(UUID()), 3, 'Помыть духовки', 1),
       (UUID_TO_BIN(UUID()), 3, 'Протереть огнетушители', 1),

       (UUID_TO_BIN(UUID()), 5, 'Пропылесосить полы', 1),
       (UUID_TO_BIN(UUID()), 5, 'Помыть полы', 1),
       (UUID_TO_BIN(UUID()), 5, 'Помыть раковину', 1),

       (UUID_TO_BIN(UUID()), 6, 'Пропылесосить ковёр в прихожей', 7),
       (UUID_TO_BIN(UUID()), 6, 'Помыть лотки для обуви (если грязные)', 5),
       (UUID_TO_BIN(UUID()), 6, 'Протереть грязь на шкафах', 1),
       (UUID_TO_BIN(UUID()), 6, 'Протереть подоконники', 1),

       (UUID_TO_BIN(UUID()), 7, 'Пропылесосить пол', 2),
       (UUID_TO_BIN(UUID()), 7, 'Помыть пол', 5),
       (UUID_TO_BIN(UUID()), 7, 'Пропылесосить коридор и лестницу', 2),
       (UUID_TO_BIN(UUID()), 7, 'Помыть пол в коридоре и лестницу', 7),
       (UUID_TO_BIN(UUID()), 7, 'Протереть поверхность подоконника', 1),
       (UUID_TO_BIN(UUID()), 7, 'Протереть поверхность машинок', 1),
       (UUID_TO_BIN(UUID()), 7, 'Запустить самоочистку у стиралок', 1),
       (UUID_TO_BIN(UUID()), 7, 'Убрать сквиш из сушилки', 2),
       (UUID_TO_BIN(UUID()), 7, 'Помыть лотки у машинок', 1),
       (UUID_TO_BIN(UUID()), 7, 'Убрать вещи за стиралкой', 1),
       (UUID_TO_BIN(UUID()), 7, 'Протереть батарею', 1),
       (UUID_TO_BIN(UUID()), 7, 'Протереть огнетушители', 1),
       (UUID_TO_BIN(UUID()), 7, 'Протереть розетки', 1);


