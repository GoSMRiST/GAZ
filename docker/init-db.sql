-- Создаём базы данных для каждого сервиса.
-- Запускается автоматически при первом старте postgres-контейнера.

CREATE DATABASE users;
CREATE DATABASE rooms;

GRANT ALL PRIVILEGES ON DATABASE users TO smrist;
GRANT ALL PRIVILEGES ON DATABASE rooms TO smrist;
