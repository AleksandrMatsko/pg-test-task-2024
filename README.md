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

    File with command will be deleted when the script finishes or if any error occurs.

- `EXECUTOR_DB_CONN_STR` - connection string which has format `postgres://{user}:{password}@{host}:{port}/{dbname}`
    
    `user`, `password`, `host`, `port`, `dbname` should be replaced with proper values.

- `EXECUTOR_MIGRATIONS_SOURCE`

# Run tests

```shell
go test -race -v ./...
```

# Some info about service

Service is used for running bash-scripts. You can interact with it by using endpoints:
- `/api/v1/cmd` - POST for uploading command, GET for listing all commands
- `/api/v1/{id}` - for getting more info about command with following id
- `/api/v1/{id}/cancel` - for canceling script execution 

## Info about endpoints

If any error occurred, server returns json (example below) and sets status code `4xx` or `5xx`
```json
{
    "short-desc": "Bad Request",
    "long-desc": "Not a shell script"
}
```

### `/api/v1/cmd`

#### Start new command

- Method: **POST**
- Request Content-Type: text/plain
- Request Body: contains script
- On success returns json (example below) and sets status code to `200`:
```json
{
    "id": "9d887cf8-7b7e-44b0-b7a6-8be72efd917a"
}
```
- On failure status codes may be: `400`, `415`, `500`

#### Get commands list

- Method: **GET**
- No Body
- On success returns json (example below) and sets status code to `200`:
```json
{
  "cmd-list": [
    {
      "id": "c5ef42d4-515a-4ec0-bb89-1a15ad522bf4",
      "status": "error",
      "status-desc": "server got down"
    },
    {
      "id": "f1e57531-b32a-4ccf-bc1d-59466682d9be",
      "status": "finished",
      "status-desc": "",
      "exit-code": 0
    },
    {
      "id": "05172c64-ba92-442f-baac-a184e545b7bf",
      "status": "running",
      "status-desc": ""
    },
    {
      "id": "f1e57531-b32a-4ccf-bc1d-59466682d9be",
      "status": "finished",
      "status-desc": "",
      "signal": 9
    }
  ]
}
```
- On failure status code: `500`

### `/api/v1/{id}`

#### Get commands info

- Method: **GET**
- No Body
- `{id}` - is a parameter returned from `POST /api/v1/cmd`
- On success returns json (examples below) and sets status code to `200`:

Example 1:
```json
{
    "id": "f1e57531-b32a-4ccf-bc1d-59466682d9be",
    "source": "#!/bin/bash\n\nls\n",
    "status": "finished",
    "status-desc": "",
    "output": "Dockerfile\nREADME.md\nbin\ndocker-compose.yaml\ngo.mod\ngo.sum\ninternal\nmain.go\npg-test-task-2024\npkg\nscripts\nsrc\ntask.md\n",
    "exit-code": 0
}
```

Example 2:
```json
{
    "id": "08783b71-4345-47d4-8e67-f91845566843",
    "source": "#!/bin/bash\n\nsleep 200\n",
    "status": "finished",
    "status-desc": "",
    "output": "",
    "signal": 9
}
```
- On failure status codes may be: `400`, `404`, `500`

### `/api/v1/{id}/cancel`

#### Cancel the command

- Method: **PATCH**
- No Body
- `{id}` - is a parameter returned from `POST /api/v1/cmd`
- On success status code is `202`
- On failure status codes may be: `400`, `404`, `500`

Trying to cancel not running command will result in 404 Not found.
