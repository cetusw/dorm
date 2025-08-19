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

INSERT INTO team (team_id, team_leader_id, color)
VALUES (2,
        UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', '')),
        252525);

UPDATE user SET team_id = 1 WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '-', ''));
UPDATE user SET team_id = 2 WHERE user_id = UNHEX(REPLACE('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', '-', ''));

INSERT INTO area (floor, name)
VALUES (3, 'Коридор');
INSERT INTO area (floor, name)
VALUES (3, 'КУИ');
INSERT INTO area (floor, name)
VALUES (1, 'Северная часть кухни');
INSERT INTO area (floor, name)
VALUES (1, 'Северный коридор');
INSERT INTO area (floor, name)
VALUES (1, 'КУИ');
INSERT INTO area (floor, name)
VALUES (1, 'Прихожая');
INSERT INTO area (floor, name)
VALUES (-1, 'Прачка');

