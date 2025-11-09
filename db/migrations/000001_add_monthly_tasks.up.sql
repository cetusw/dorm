INSERT IGNORE INTO `task` (`task_id`, `area_id`, `task_title`, `task_cost`, `task_frequency`)
VALUES (UUID_TO_BIN(UUID()), 5, 'Помыть холодильники внутри', 9, 30),
       (UUID_TO_BIN(UUID()), 8, 'Помыть холодильники внутри', 9, 30);