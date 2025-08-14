USE dorm;
CREATE TABLE user
(
    user_id      BINARY(16)    NOT NULL,
    telegram_id  BIGINT UNIQUE NOT NULL,
    first_name   VARCHAR(100)  NOT NULL,
    last_name    VARCHAR(100)  NOT NULL,
    middle_name  VARCHAR(100) DEFAULT '',
    team_id      BINARY(16)   DEFAULT 0,
    room_number  INT          DEFAULT 0,
    dormitory_id BINARY(16)   DEFAULT 0,
    created_at   TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMP    DEFAULT NULL
);