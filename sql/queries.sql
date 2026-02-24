-- name: CreateSubscription :one
INSERT INTO subscriptions (
    service_name,
    price,
    user_uuid,
    start_date,
    end_date    
) VALUES (
    @service_name,
    @price,
    @user_uuid,
    @start_date,
    @end_date    
) RETURNING *;

-- name: GetSubscriptionById :one
SELECT * FROM subscriptions
WHERE id = @id;

-- name: GetAllSubscriptions :many
SELECT * FROM subscriptions
WHERE
    user_uuid = COALESCE(sqlc.narg('user_uuid'), user_uuid) AND
    service_name = COALESCE(sqlc.narg('service_name'), service_name);

-- name: CalculateSubscriptionsRevenue :one
SELECT
    COALESCE(
        SUM(
            price * (EXTRACT(MONTH FROM overlap_end) - EXTRACT(MONTH FROM overlap_start) +
            (EXTRACT(YEAR FROM overlap_end) - EXTRACT(YEAR FROM overlap_start)) * 12 + 1)
        ),
    0) :: BIGINT
FROM (
    SELECT
        price,
        LEAST(@period_end, cancelation_date, end_date) AS overlap_end,
        GREATEST(@period_start, start_date) AS overlap_start
    FROM subscriptions
    WHERE
        tsrange(
            sqlc.narg(period_start)::DATE,
            sqlc.narg(period_end)::DATE + INTERVAL '1 month'
        ) && tsrange(
            start_date,
            end_date + INTERVAL '1 month'
        ) AND
        user_uuid = COALESCE(sqlc.narg('user_uuid'), user_uuid) AND
        service_name = COALESCE(sqlc.narg('service_name'), service_name) AND
        COALESCE(cancelation_date >= start_date, TRUE)
);

-- name: CancelSubscriptionByID :execrows
UPDATE subscriptions SET
    cancelation_date = COALESCE(cancelation_date, @cancelation_date)
WHERE
    id = @id AND
    COALESCE(end_date >= @cancelation_date, TRUE);
