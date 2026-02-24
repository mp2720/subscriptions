-- +goose Up
-- +goose StatementBegin
CREATE INDEX subscriptions_user_uuid_idx ON subscriptions(user_uuid);
CREATE INDEX subscriptions_service_name_idx ON subscriptions(service_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX subscriptions_user_uuid_idx;
DROP INDEX subscriptions_service_name_idx;
-- +goose StatementEnd
