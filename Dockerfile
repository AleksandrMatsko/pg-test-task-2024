FROM golang:1.22
LABEL authors="AMats"

COPY . .

RUN go build -v pg-test-task-2024

ENTRYPOINT ["./pg-test-task-2024"]
