# SP-Share

### Start the web server:

```bash
$ revel run -a sp-share
```

### Use local postgres database using Docker

```bash
# Start the postgres container 
$ docker run --rm --name testdb -e POSTGRES_PASSWORD=docker -e POSTGRES_DB=spdb -d -p 5432:5432 postgres:11.6

# Check the status of container
$ docker ps
CONTAINER ID        IMAGE               COMMAND                  CREATED              STATUS              PORTS                    NAMES
642ed8e81af4        postgres:11.6       "docker-entrypoint.s…"   About a minute ago   Up About a minute   0.0.0.0:5432->5432/tcp   testdb
```
To access the database run the following command

```bash
$ psql -h localhost -U postgres -d spdb
```
