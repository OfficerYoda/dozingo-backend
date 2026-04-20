-- Drop triggers
DROP TRIGGER IF EXISTS set_updated_at ON board_votes;
DROP TRIGGER IF EXISTS set_updated_at ON bingo_cells;
DROP TRIGGER IF EXISTS set_updated_at ON bingo_boards;
DROP TRIGGER IF EXISTS set_updated_at ON lecturers;
DROP TRIGGER IF EXISTS set_updated_at ON user_authentications;
DROP TRIGGER IF EXISTS set_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS set_updated_at();

-- Drop tables (reverse order to respect foreign keys)
DROP TABLE IF EXISTS board_votes;
DROP TABLE IF EXISTS bingo_cells;
DROP TABLE IF EXISTS bingo_boards;
DROP TABLE IF EXISTS lecturers;
DROP TABLE IF EXISTS user_authentications;
DROP TABLE IF EXISTS users;

-- Drop extension
DROP EXTENSION IF EXISTS "pgcrypto";
