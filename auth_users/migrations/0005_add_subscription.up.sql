-- Подписки: default (бесплатная), vip, super
-- Хранится прямо в users — при выходе отдельного subscription-сервиса
-- это поле станет кешем или будет убрано в пользу gRPC-вызова.
ALTER TABLE users
    ADD COLUMN subscription TEXT NOT NULL DEFAULT 'default'
        CHECK (subscription IN ('default', 'vip', 'super'));
