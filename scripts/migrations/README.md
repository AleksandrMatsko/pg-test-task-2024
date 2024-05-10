# Migrations guide

For managing migrations this [tool](https://github.com/golang-migrate/migrate) is used. 
The installation guide may be found here [CLI](https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md)

## Apply migrations

To apply migrations use such command:

```shell
migrate -database "postgres://$user:$password@$host:$port/$dbname?sslmode=disable" -path $dirname up
```
The following variables should be replaced with actual values:
- `$user`
- `$password`
- `$host`
- `$port`
- `$dbname`
- `$dirname` - may be absolute or relative

## Create new migration

```shell
migrate create -ext sql -dir $dirname -seq $migration_name
```

Replace:
- `$dirname` - may be absolute or relative
- `$migration_name`

with proper values

