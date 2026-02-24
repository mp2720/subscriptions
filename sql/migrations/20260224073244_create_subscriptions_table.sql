-- +goose Up
-- +goose StatementBegin
CREATE TABLE subscriptions (
    id           BIGINT GENERATED ALWAYS AS IDENTITY,
    service_name TEXT NOT NULL,
    price        INT NOT NULL,
    user_uuid    BYTEA NOT NULL,
    start_date   DATE NOT NULL,
    end_date     DATE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE subscriptions;
-- +goose StatementEnd
