-- создание ролей

INSERT INTO role (role_id, role_name, role_description)
VALUES (1, 'Житель', 'Живёт в коливинге');

-- создание общежитий

INSERT INTO dormitory (dormitory_id, dormitory_name, dormitory_city, dormitory_street_type, dormitory_street_name,
                       dormitory_house_number)
VALUES (1, 'Завод', 'Йошкар-Ола', 'переулок', 'Заводской', '3А'),
       (2, 'Гагарин', 'Йошкар-Ола', 'улица', 'Волкова', '108');

-- создание моего пользователя

INSERT INTO user (user_id, telegram_id, first_name, last_name, role_id)
VALUES (UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', '')), 523562011, 'Михаил', 'Кугелев', 1);

-- создание тестовых пользователей

INSERT INTO user (user_id, telegram_id, first_name, last_name, role_id)
VALUES (UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', '')), 523562012, 'Тест', '1', 1),
       (UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', '-', '')), 523562013, 'Тест', '2', 1),
       (UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', '-', '')), 523562014, 'Тест', '3', 1),
       (UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '-', '')), 523562015, 'Тест', '4', 1),
       (UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', '-', '')), 523562016, 'Тест', '5', 1),
       (UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', '-', '')), 523562017, 'Тест', '6', 1),
       (UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a18', '-', '')), 523562018, 'Тест', '2', 1),
       (UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a19', '-', '')), 523562019, 'Тест', '2', 1),
       (UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a20', '-', '')), 523562020, 'Тест', '2', 1),
       (UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', '-', '')), 523562021, 'Тест', '10', 1);

-- создание групп команд

INSERT INTO `group` (group_id, dormitory_id, group_name, spreadsheet_id)
VALUES (UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', '')), 1, 'Мужчины',
        '1fYSv20bhoiHZlQY7TVR-0OfK8EUKmPsCaoQpEJKxMgo'),
       (UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', '')), 1, 'Женщины',
        '1SHtbP1YZvsGEfI-o_JKnk-5NlaLHav_qjyAqKsIuaIY');

-- создание команд

INSERT INTO team (team_id, group_id, team_leader_id, team_color, team_order)
VALUES (UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', '')),
        UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', '')),
        UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', '')), 'b6d7a8', 1),
       (UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', '')),
        UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', '')),
        UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', '')), 'b4a7d6', 2),
       (UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', '-', '')),
        UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', '')),
        UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', '-', '')), 'fff2cc', 1),
       (UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', '-', '')),
        UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', '')),
        UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', '-', '')), 'e6b8af', 2);

-- распределение пользователей по командам

UPDATE user
SET team_id = UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', ''))
WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', ''));
UPDATE user
SET team_id = UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', ''))
WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', '-', ''));
UPDATE user
SET team_id = UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', ''))
WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a19', '-', ''));
UPDATE user
SET team_id = UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', ''))
WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', ''));
UPDATE user
SET team_id = UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', ''))
WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a16', '-', ''));
UPDATE user
SET team_id = UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', ''))
WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a21', '-', ''));
UPDATE user
SET team_id = UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', '-', ''))
WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', '-', ''));
UPDATE user
SET team_id = UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', '-', ''))
WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a17', '-', ''));
UPDATE user
SET team_id = UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', '-', ''))
WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a20', '-', ''));
UPDATE user
SET team_id = UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', '-', ''))
WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', '-', ''));
UPDATE user
SET team_id = UNHEX(REPLACE('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', '-', ''))
WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a18', '-', ''));

-- создание зон дежурства

INSERT INTO area (area_id, area_floor, area_name, group_id)
VALUES (1, 3, 'Коридор', UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', ''))),
       (2, 3, 'КУИ', UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', ''))),
       (3, 2, 'Коридор', UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', ''))),
       (4, 2, 'КУИ', UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', ''))),
       (5, 1, 'Северная часть кухни', UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', ''))),
       (6, 1, 'Северный коридор', UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', ''))),
       (7, 1, 'КУИ', NULL),
       (8, 1, 'Южная часть кухни', UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', ''))),
       (9, 1, 'Южный коридор', UNHEX(REPLACE('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', ''))),
       (10, 1, 'Прихожая', NULL),
       (11, -1, 'Прачка', NULL)
;

-- создание задач

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

       (UUID_TO_BIN(UUID()), 3, 'Пропылесосить полы', 2),
       (UUID_TO_BIN(UUID()), 3, 'Помыть полы', 8),
       (UUID_TO_BIN(UUID()), 3, 'Протереть плинтуса', 3),
       (UUID_TO_BIN(UUID()), 3, 'Разобрать старые вещи на сушилках', 1),
       (UUID_TO_BIN(UUID()), 3, 'Протереть батареи', 1),
       (UUID_TO_BIN(UUID()), 3, 'Протереть подоконники', 2),
       (UUID_TO_BIN(UUID()), 3, 'Протереть окна', 4),
       (UUID_TO_BIN(UUID()), 3, 'Протереть двери', 2),
       (UUID_TO_BIN(UUID()), 3, 'Протереть огнетушители', 1),

       (UUID_TO_BIN(UUID()), 4, 'Пропылесосить полы', 1),
       (UUID_TO_BIN(UUID()), 4, 'Помыть полы', 1),
       (UUID_TO_BIN(UUID()), 4, 'Протереть подоконник', 1),
       (UUID_TO_BIN(UUID()), 4, 'Помыть раковину', 1),

       (UUID_TO_BIN(UUID()), 5, 'Пропылесосить полы', 2),
       (UUID_TO_BIN(UUID()), 5, 'Помыть полы', 6),
       (UUID_TO_BIN(UUID()), 5, 'Протереть батареи', 2),
       (UUID_TO_BIN(UUID()), 5, 'Протереть подоконники', 2),
       (UUID_TO_BIN(UUID()), 5, 'Протереть розетки и выключатели', 1),
       (UUID_TO_BIN(UUID()), 5, 'Убрать мусор из раковины и помыть её', 3),
       (UUID_TO_BIN(UUID()), 5, 'Собрать и вынести мусор + помыть место под мусоркой', 2),
       (UUID_TO_BIN(UUID()), 5, 'Протереть столы', 1),
       (UUID_TO_BIN(UUID()), 5, 'Помыть кухонный гарнитур', 3),
       (UUID_TO_BIN(UUID()), 5, 'Разобрать посуду', 1),
       (UUID_TO_BIN(UUID()), 5, 'Помыть ёмкости для столовых приборов', 1),
       (UUID_TO_BIN(UUID()), 5, 'Протереть полки в шкафчиках(где грязно)', 1),
       (UUID_TO_BIN(UUID()), 5, 'Постирать тряпки', 1),
       (UUID_TO_BIN(UUID()), 5, 'Помыть холодильники снаружи', 3),
       (UUID_TO_BIN(UUID()), 5, 'Избавиться от просрочки в холодильниках', 3),
       (UUID_TO_BIN(UUID()), 5, 'Помыть духовки', 1),
       (UUID_TO_BIN(UUID()), 5, 'Протереть огнетушители', 1),

       (UUID_TO_BIN(UUID()), 6, 'Пропылесосить полы', 2),
       (UUID_TO_BIN(UUID()), 6, 'Помыть полы', 6),
       (UUID_TO_BIN(UUID()), 6, 'Протереть плинтуса', 3),
       (UUID_TO_BIN(UUID()), 6, 'Протереть батареи', 2),
       (UUID_TO_BIN(UUID()), 6, 'Протереть розетки и выключатели', 1),
       (UUID_TO_BIN(UUID()), 6, 'Помыть северную лестницу', 9),
       (UUID_TO_BIN(UUID()), 6, 'Протереть огнетушители', 1), -- У задачи не было баллов, я поставил 1

       (UUID_TO_BIN(UUID()), 7, 'Пропылесосить полы', 1),
       (UUID_TO_BIN(UUID()), 7, 'Помыть полы', 1),
       (UUID_TO_BIN(UUID()), 7, 'Помыть раковину', 1),

       (UUID_TO_BIN(UUID()), 8, 'Пропылесосить полы', 2),
       (UUID_TO_BIN(UUID()), 8, 'Помыть полы', 6),
       (UUID_TO_BIN(UUID()), 8, 'Протереть батареи', 2),
       (UUID_TO_BIN(UUID()), 8, 'Протереть подоконники', 2),
       (UUID_TO_BIN(UUID()), 8, 'Протереть розетки и выключатели', 1),
       (UUID_TO_BIN(UUID()), 8, 'Убрать мусор из раковины и помыть её', 3),
       (UUID_TO_BIN(UUID()), 8, 'Собрать и вынести мусор + помыть место под мусоркой', 2),
       (UUID_TO_BIN(UUID()), 8, 'Протереть столы', 1),
       (UUID_TO_BIN(UUID()), 8, 'Помыть кухонный гарнитур', 3),
       (UUID_TO_BIN(UUID()), 8, 'Разобрать посуду', 1),
       (UUID_TO_BIN(UUID()), 8, 'Помыть ёмкости для столовых приборов', 1),
       (UUID_TO_BIN(UUID()), 8, 'Протереть полки в шкафчиках(где грязно)', 1),
       (UUID_TO_BIN(UUID()), 8, 'Постирать тряпки', 1),
       (UUID_TO_BIN(UUID()), 8, 'Помыть холодильники снаружи', 3),
       (UUID_TO_BIN(UUID()), 8, 'Избавиться от просрочки в холодильниках', 3),
       (UUID_TO_BIN(UUID()), 8, 'Помыть духовки', 1),
       (UUID_TO_BIN(UUID()), 8, 'Протереть огнетушители', 1),

       (UUID_TO_BIN(UUID()), 9, 'Пропылесосить полы', 2),
       (UUID_TO_BIN(UUID()), 9, 'Помыть полы', 6),
       (UUID_TO_BIN(UUID()), 9, 'Протереть плинтуса', 3),
       (UUID_TO_BIN(UUID()), 9, 'Протереть батареи', 2),
       (UUID_TO_BIN(UUID()), 9, 'Протереть розетки и выключатели', 1),
       (UUID_TO_BIN(UUID()), 9, 'Помыть южную лестницу', 9),
       (UUID_TO_BIN(UUID()), 9, 'Протереть огнетушители', 1),

       (UUID_TO_BIN(UUID()), 10, 'Пропылесосить ковёр в прихожей', 7),
       (UUID_TO_BIN(UUID()), 10, 'Помыть лотки для обуви (если грязные)', 5),
       (UUID_TO_BIN(UUID()), 10, 'Протереть грязь на шкафах', 1),
       (UUID_TO_BIN(UUID()), 10, 'Протереть подоконники', 1),

       (UUID_TO_BIN(UUID()), 11, 'Пропылесосить пол', 2),
       (UUID_TO_BIN(UUID()), 11, 'Помыть пол', 5),
       (UUID_TO_BIN(UUID()), 11, 'Пропылесосить коридор и лестницу', 2),
       (UUID_TO_BIN(UUID()), 11, 'Помыть пол в коридоре и лестницу', 7),
       (UUID_TO_BIN(UUID()), 11, 'Протереть поверхность подоконника', 1),
       (UUID_TO_BIN(UUID()), 11, 'Протереть поверхность машинок', 1),
       (UUID_TO_BIN(UUID()), 11, 'Запустить самоочистку у стиралок', 1),
       (UUID_TO_BIN(UUID()), 11, 'Убрать сквиш из сушилки', 2),
       (UUID_TO_BIN(UUID()), 11, 'Помыть лотки у машинок', 1),
       (UUID_TO_BIN(UUID()), 11, 'Убрать вещи за стиралкой', 1),
       (UUID_TO_BIN(UUID()), 11, 'Протереть батарею', 1),
       (UUID_TO_BIN(UUID()), 11, 'Протереть огнетушители', 1),
       (UUID_TO_BIN(UUID()), 11, 'Протереть розетки', 1);
