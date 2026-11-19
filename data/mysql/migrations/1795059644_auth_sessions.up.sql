CREATE TEMPORARY TABLE auth_login_preflight (ok INT NOT NULL CONSTRAINT active_login_conflicts_require_manual_resolution CHECK (ok=1));
INSERT INTO auth_login_preflight SELECT 0 FROM `user` WHERE deleted_at IS NULL GROUP BY login HAVING COUNT(*)>1;
DROP TEMPORARY TABLE auth_login_preflight;

CREATE TABLE auth_session
(
    id         BINARY(32)   NOT NULL PRIMARY KEY,
    user_id    BINARY(16)   NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    expires_at TIMESTAMP(6) NOT NULL,
    revoked_at TIMESTAMP(6) NULL,
    INDEX ix_auth_session_user (user_id),
    INDEX ix_auth_session_expiry (expires_at),
    CONSTRAINT fk_auth_session_user FOREIGN KEY (user_id) REFERENCES `user` (id)
) ENGINE = InnoDB;

CREATE UNIQUE INDEX uq_active_user_login ON `user` ((CASE WHEN deleted_at IS NULL THEN login ELSE NULL END));
