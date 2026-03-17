FROM golang:1.26.1

RUN mkdir "/app"

WORKDIR /app

CMD go mod tidy && go run /app/cmd/server/main.go
