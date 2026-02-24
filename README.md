[Задание](https://disk.360.yandex.ru/i/rtGuSd4l-2a-Fw)

## Docker

Пример `.env`:

```
DB=subscriptions
DB_USER=user
DB_PASSWORD=password
DB_PORT=5432
DB_VOLUME=/tmp/postgres-data
SERVICE_VERBOSE=true
SERVICE_PORT=8080
```

```bash
docker build -t subscriptions .
docker compose up
```

## Генерация кода

Сгенерированный код уже есть в репозитории, эти команды можно не выполнять.

```sh
swag init
sqlc generate -f sql/sqlc.yaml
```

`sqlc` требуется подключение к БД ([managed databases](https://docs.sqlc.dev/en/latest/howto/managed-databases.html)),
URL можно задать в `sqlc/sqlc.yaml`

## Запуск без docker

Нужно задать следующие env переменные:

```
DB=subscriptions
DB_USER=user
DB_PASSWORD=password
DB_HOST=localhost
DB_PORT=5432
DB_SSL_MODE=disable
SERVICE_VERBOSE=true
SERVICE_PORT=8080
```

Если они записаны в `.env`, можно экспортировать так (работает в `bash` и `fish`):

```bash
eval export $(cat .env)
```

Запуск

```bash
go run .
```

## Swagger

URL: `http://host:port/swagger/index.html`

## Примечания

В моей интерпретации задания оплата подписки происходит каждый месяц, пока подписка активна.

Операция удаления, как мне кажется, не имеет смысла.
Вместо этого я сделал отмену.
Дата отмены сохраняется и учитывается при подсчёте суммы.
Отменить можно только подписки, которые до сих пор могут быть активны.

