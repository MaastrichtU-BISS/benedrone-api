# Graphhopper api

## Starting/stopping

To start:
```bash
docker compose up -d
```

To stop:
```bash
docker compose down
```

To show logs of running container:
```bash
docker compose logs api
```

To show real-time logs of container
```bash
docker compose logs -f api
```

To stop real-time monitoring of logs in the last command, type ctrl+c


## To run the jar buil
java -jar graphhopper-web-11.0.jar server config.yml

