ALTER TABLE penalty_entry
    MODIFY reason VARCHAR(512) NOT NULL;

CREATE TABLE individual_task (
    id BINARY(16) NOT NULL,
    dormitory_id INT UNSIGNED NOT NULL,
    resident_id BINARY(16) NOT NULL,
    area_id INT UNSIGNED NULL,
    title VARCHAR(255) NOT NULL,
    redemption_weight DECIMAL(10,1) NOT NULL DEFAULT 0,
    status ENUM('ISSUED','COMPLETED','VERIFIED') NOT NULL DEFAULT 'ISSUED',
    deadline DATE NULL,
    completed_at DATETIME(6) NULL,
    verified_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    deleted_at DATETIME(6) NULL,
    version BIGINT UNSIGNED NOT NULL DEFAULT 1,
    PRIMARY KEY (id),
    KEY idx_individual_task_resident_status_deleted (resident_id, status, deleted_at),
    KEY idx_individual_task_dormitory_status_deleted (dormitory_id, status, deleted_at),
    CONSTRAINT chk_individual_task_redemption_weight CHECK (redemption_weight >= 0),
    CONSTRAINT fk_individual_task_dormitory FOREIGN KEY (dormitory_id) REFERENCES dormitory(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_individual_task_resident FOREIGN KEY (resident_id) REFERENCES user(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_individual_task_area FOREIGN KEY (area_id) REFERENCES area(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
