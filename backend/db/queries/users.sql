-- ============================================================
-- USERS
-- ============================================================

-- name: CreateUser :one
WITH inserted_user AS (
    INSERT INTO users (name, email, password_hash)
    VALUES ($1, $2, $3)
    RETURNING id, name, email, avatar_url, timezone, email_verified, created_at, updated_at
),
inserted_roles AS (
    INSERT INTO user_roles (user_id, role_id)
    SELECT iu.id, r.role_id
    FROM inserted_user iu
    CROSS JOIN unnest(sqlc.arg('role_ids')::int[]) AS r(role_id)
    RETURNING user_id
)
SELECT iu.*
FROM inserted_user iu
JOIN inserted_roles ir ON ir.user_id = iu.id
GROUP BY iu.id, iu.name, iu.email, iu.avatar_url, iu.timezone, iu.email_verified, iu.created_at, iu.updated_at;

-- name: GetUserByID :one
SELECT
    u.id, u.name, u.email, u.avatar_url, u.timezone, u.email_verified,
    u.created_at, u.updated_at,
    array_agg(ur.role_id ORDER BY ur.role_id)::int[] AS role_ids
FROM users u
LEFT JOIN user_roles ur ON ur.user_id = u.id
WHERE u.id = $1
GROUP BY u.id;

-- name: GetUserByEmail :one
SELECT
    u.id, u.name, u.email, u.password_hash, u.avatar_url, u.timezone, u.email_verified,
    u.created_at, u.updated_at,
    array_agg(ur.role_id ORDER BY ur.role_id)::int[] AS role_ids
FROM users u
LEFT JOIN user_roles ur ON ur.user_id = u.id
WHERE u.email = $1
GROUP BY u.id;

-- name: UpdateUser :one
WITH updated AS (
    UPDATE users
    SET name       = COALESCE(sqlc.narg('name'), name),
        avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
        timezone   = COALESCE(sqlc.narg('timezone'), timezone),
        updated_at = NOW()
    WHERE id = sqlc.arg('id')
    RETURNING id, name, email, avatar_url, timezone, email_verified, created_at, updated_at
)
SELECT
    u.id, u.name, u.email, u.avatar_url, u.timezone, u.email_verified,
    u.created_at, u.updated_at,
    array_agg(ur.role_id ORDER BY ur.role_id)::int[] AS role_ids
FROM updated u
LEFT JOIN user_roles ur ON ur.user_id = u.id
GROUP BY u.id, u.name, u.email, u.avatar_url, u.timezone, u.email_verified, u.created_at, u.updated_at;

-- name: SetEmailVerified :exec
UPDATE users SET email_verified = TRUE, updated_at = NOW() WHERE id = $1 AND email_verified = FALSE;

-- name: UpdatePassword :exec
UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- ============================================================
-- ROLES
-- ============================================================

-- name: AssignRole :exec
INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveRole :exec
DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2;

-- name: GetUserRoles :many
SELECT role_id FROM user_roles WHERE user_id = $1;

-- name: SetUserRoles :exec
DELETE FROM user_roles WHERE user_id = $1;

-- name: AddUserRoles :exec
INSERT INTO user_roles (user_id, role_id)
SELECT $1, unnest($2::int[])
ON CONFLICT DO NOTHING;

-- name: HasRole :one
SELECT EXISTS(
    SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2
) AS has_role;

-- ============================================================
-- EMAIL VERIFICATION
-- ============================================================

-- name: CreateEmailVerificationToken :one
INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, expires_at, created_at;

-- name: GetEmailVerificationToken :one
SELECT id, user_id, expires_at, used_at
FROM email_verification_tokens
WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW();

-- name: MarkEmailVerificationTokenUsed :exec
UPDATE email_verification_tokens SET used_at = NOW() WHERE id = $1;

-- ============================================================
-- PASSWORD RESET
-- ============================================================

-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, expires_at, created_at;

-- name: GetPasswordResetToken :one
SELECT id, user_id, expires_at, used_at
FROM password_reset_tokens
WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW();

-- name: MarkPasswordResetTokenUsed :exec
UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1;

-- name: DeleteExpiredPasswordResetTokens :exec
DELETE FROM password_reset_tokens WHERE expires_at < NOW();