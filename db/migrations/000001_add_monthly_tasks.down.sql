DELETE FROM `task`
WHERE `task_title` = 'Помыть холодильники внутри'
  AND `area_id` IN (5, 8);