DROP TRIGGER IF EXISTS trg_user_roles_updated_at ON user_roles;
DROP TRIGGER IF EXISTS trg_roles_updated_at ON roles;

DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;