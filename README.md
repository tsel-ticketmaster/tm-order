# tm-order
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=tsel-ticketmaster_tm-order&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=tsel-ticketmaster_tm-order)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=tsel-ticketmaster_tm-order&metric=bugs)](https://sonarcloud.io/summary/new_code?id=tsel-ticketmaster_tm-order)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=tsel-ticketmaster_tm-order&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=tsel-ticketmaster_tm-order)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=tsel-ticketmaster_tm-order&metric=coverage)](https://sonarcloud.io/summary/new_code?id=tsel-ticketmaster_tm-order)
[![Duplicated Lines (%)](https://sonarcloud.io/api/project_badges/measure?project=tsel-ticketmaster_tm-order&metric=duplicated_lines_density)](https://sonarcloud.io/summary/new_code?id=tsel-ticketmaster_tm-order)


This Project is used to handle ticket master customer services including auth, registration and profile.

### Prerequisites

What things you need to install the software and how to install them

```
Golang v1.21.x
Go Mod
....
```

### Installing

A step by step series of examples that tell you have to get a development env running

Say what the step will be
- Create ENV file (.env) with this configuration:
```
APP_NAME=tm-order
APP_PORT=9000
APP_ENVIRONMENT=dev
APP_TIMEZONE=Asia/Jakarta
APP_DEBUG=TRUE
APP_TIMEOUT=2
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=OPTIONS,POST,GET,PUT,PATCH,DELETE
CORS_ALLOWED_HEADERS=
CORS_EXPOSED_HEADERS=
CORS_ALLOW_CREDENTIALS=TRUE
CORS_MAX_AGE=5
REDIS_HOSTS=localhost:6379
REDIS_PASSWORD=redispass
REDIS_DB=0
POSTGRESQL_HOST=localhost
POSTGRESQL_PORT=5432
POSTGRESQL_USER=patrick
POSTGRESQL_PASSWORD=12345678
POSTGRESQL_DBNAME=ticket-master
POSTGRESQL_SSLMODE=disable
POSTGRESQL_MAX_OPEN_CONNS=100
POSTGRESQL_MAX_IDLE_CONNS=100
JWT_RSA=
```
- Then run this command (Development Issues)
```
Give the example
...
$ make run.dev
```

- Then run this command (Production Issues)
```
Give the example
...
$ make install
$ make test
$ make build
$ ./app
```

### Running the tests

Explain how to run the automated tests for this system
```sh
Give the example
...
$ make test
```

### Running the tests (With coverage appear on)

Explain how to run the automated tests for this system
```sh
Give the example
...
$ make cover
```

### Deployment

Add additional notes about how to deploy this on a live system

### Built With

* [Gorilla/Mux](https://github.com/gorilla/mux) The rest framework used
* [Mockery] Mock Up Generator
* [GoMod] - Dependency Management
* [Docker] - Container Management

### Authors

* **Patrick Maurits Sangian** - [Github](https://github.com/sangianpatrick)
