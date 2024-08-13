FROM golang:1.22.1

WORKDIR /app

COPY /back ./
COPY .env /.env

RUN go mod tidy