# Start server

## Preferred way

Use docker compose:
```shell
docker compose up
```

## Other

If you need to run the service outside the container please use commands below and 
pay attention to configuration.

```shell
go build -v pg-test-task-2024
./pg-test-task-2024
```

# Configuration

Service is configured by setting some environment variables:
- `EXECUTOR_HOST` - host to listen
- `EXECUTOR_PORT` - port to listen
- `EXECUTOR_CMD_DIR` - path to directory, where files with commands will be stored. 

    File with command will be deleted when the script will be finished or if any error occurred.

- `EXECUTOR_DB_CONN_STR` - connection string which has format `postgres://{user}:{password}@{host}:{port}/{dbname}`
    
    `user`, `password`, `host`, `port`, `dbname` should be replaced with proper values.

- `EXECUTOR_MIGRATIONS_SOURCE`

# Run tests

```shell
go test -race -v ./...
```

# Some info about service

Service is used for running bash-scripts. You can interact with it by using endpoints:
- `/api/v1/cmd` - POST for uploading command, GET for listing all command
- `/api/v1/{id}` - for getting more info about command with following id
- `/api/v1/{id}/cancel` - for canceling script execution 
