FROM golang:1.25

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN ls /usr/local/bin
RUN go build -v -o /usr/local/bin/ ./...

CMD ["/usr/local/bin/subscriptions"]
