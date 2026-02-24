-- +goose Up
-- +goose StatementBegin
ALTER TABLE subscriptions ADD cancelation_date DATE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE subscriptions DROP COLUMN cancelation_date;
-- +goose StatementEnd
