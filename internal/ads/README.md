# Ads Manager

The **Ads Manager** service for the litit.

## Setup

### Install dependencies

**Note**: Do not run the make commands in stage or production environment.

```shell
make setup
```

Copy the configurations from the `config.json` to `config.qwerty.json` and change the values accordingly.
```shell
cp config.json config.qwerty.json
```

### Run the server

```shell
make run
```

Alternatively, you can run the server using docker.
### Run using Docker

```shell
make docker-run
```

## Database

Run the migrations to create the tables in the database.

```shell
make migrations
```
