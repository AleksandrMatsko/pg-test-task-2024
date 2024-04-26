FROM golang:1.22
LABEL authors="AMats"

COPY . .

RUN go build -v ./...

ENTRYPOINT ["./pg-test-task-2024"]
