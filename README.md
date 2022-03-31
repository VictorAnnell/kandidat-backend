# Kandidat Backend

## Usage quick start

A working [Go][] installation is required. [Docker][] and [Docker Compose][] are (recommended) optional dependencies for setting up a local PostgresSQL database instance. Provided these prerequisites are in place the project can be bootstrapped and run with the following commands:

```shell
git clone 'https://github.com/VictorAnnell/kandidat-backend'
cd kandidat-backend
docker-compose up -d
go run main.go
```

## Configuration parameters

While the project uses sane defaults a range of runtime parameters can be set if needed. These are read from environment variables or a `.env` file, with the former having priority if both are present. The file [.env.sample](.env.sample) contains a list of recognized parameters along with their default values.

## Developing

### Setup local PostgresSQL database with docker

While any PostgresSQL database can be used it's recommended to use the provided docker-compose.yml file for easy setup of a local PostgresSQL database to develop and test against. The database will be prepopulated with tables defined in the file [db/init.sql](db/init.sql). The following steps are needed to get started with using docker:

1. Install [Docker][] and [Docker Compose][].

2. Clone this git repository if you haven't already and navigate to the repository directory:

```shell
git clone 'https://github.com/VictorAnnell/kandidat-backend'
cd kandidat-backend
```

3. Use the following commands to manage the docker containers:

Create and start all containers defined in the local docker-compose.yml file in the background:

```
docker-compose up -d
```

List all running containers in the current project:

```
docker-compose ps
```

Stop and remove all containers and volumes in the current project:

```
docker-compose down -v
```

> Note: The above command followed by a subsequent `docker-compose up -d` needs to be done every time you want the database to be recreated using the definitions in [db/init.sql](db/init.sql).

Print the logs of the containers in the current project:

```
docker-compose logs
```

#### Adminer

Along with setting up a local instance of PostgresSQL the provided [docker-compose.yml](docker-compose.yml) file defines a Adminer container which is a web interface you can use to administer the database if needed.
While the containers are running the web interface can be accessed in a browser at the address [`localhost:8081`](http://localhost:8081). There you can use the following credentials to login:

```
Server: db
Username: dbuser
Password: kandidat-backend
Database: backend-db
```

### Building, running and maintaining the codebase

Provided you have [Go][] installed you can use the following commands from the project directory to:

Install required dependencies, build and run project:

```shell
go run main.go
```

Run tests:

```
go test
```

Build executable:

```
go build
```

> Note: The actions mentioned below should be performed before a commit/pull request.

Reformat source code:

```
go fmt
```

Verify dependencies:

```
go mod verify
```

Add missing and remove unused modules:

```
go mod tidy
```

#### Linting

Along with running the test suite the project also uses [golangci-lint][] to vet pull requests and pushes to the main branch. It is therefore recommended to [install golangci-lint locally][] and run it before any commit/push. With golangci-lint installed this is done with the following command from the project directory:

```
golangci-lint run
```

[docker compose]: https://docs.docker.com/compose/install/
[docker]: https://www.docker.com/
[go]: https://go.dev/
[golangci-lint]: https://github.com/golangci/golangci-lint
[install golangci-lint locally]: https://golangci-lint.run/usage/install/#local-installation
