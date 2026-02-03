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

## maps/
Create a maps/ directory in the root. Copy an [osm file](https://download.geofabrik.de/europe/netherlands.html) there. We recommend to use [osmium](https://osmcode.org/osmium-tool/) to merge Belgium and Dutch maps into a single one called "bene.osm.pbf".

## graph-cache/
Depending on the size of the map that is used, the initialization could take several minutes. Then the resulting graphs will be stored in /graph-cache to be reused from that point on.


