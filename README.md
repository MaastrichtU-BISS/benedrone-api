# BeNeDrone routing API (GraphHopper)

GraphHopper 11.0 serving the car, bike and foot modes for the BeNeDrone client
over the Netherlands, Belgium and Germany. Drone routes do not come from here -
they come from the motion-planner's aviation grid.

## The token gate

GraphHopper's open-source server has no authentication of its own; the
Dropwizard auth module is not even bundled in the jar. So the check lives in
front of it: **Caddy is the only published container**, it rejects any request
without a matching `X-API-Key`, and GraphHopper is reachable solely on the
internal compose network.

```
client -> Vercel Function (adds the token) -> Caddy -> GraphHopper
```

`/health` is the one unauthenticated route, and it proxies through to
GraphHopper's own health resource rather than being answered by Caddy - so a
green check means the whole chain works, not just that the gate is listening.

The token is `GRAPHHOPPER_TOKEN` in the stack's environment. It is never
committed, and the matching value lives in Vercel as the client's secret.

## Running

```bash
docker compose up -d          # start
docker compose down           # stop
docker compose logs -f caddy  # follow the gate
```

Requests go to Caddy, not to GraphHopper:

```bash
curl -H "X-API-Key: $GRAPHHOPPER_TOKEN" \
  "http://localhost:8080/route?point=52.5,5.8&point=51.9,6.6&profile=car"
```

Running the jar directly, without the gate:

```bash
java -jar graphhopper-web-11.0.jar server config.yml
```

## maps/ and graph-cache/

Neither is in the repository - they are gigabytes, and the cache is derived.

`maps/` holds the OSM extract named by `datareader.file` in `config.yml`:
`benege.osm.pbf`. Download the [Geofabrik](https://download.geofabrik.de/europe.html)
extracts for the three countries and merge them with
[osmium](https://osmcode.org/osmium-tool/).

`graph-cache/` is what GraphHopper builds from it on first start, which takes
several minutes at this size and is reused from then on. The cache is ~4.8 GB
and is memory-mapped, which is why the healthcheck's `start_period` is generous:
the first request has to page it in off disk.

## Deploying to Coolify

Three constraints, each learned the hard way and each already encoded in the
files - they are listed here so a future change does not undo them.

**Configuration is baked into the image, not mounted.** Coolify rewrites bind
mount sources into its own managed storage directory rather than the repo
clone, so a mounted `config.yml` was never where the container looked and it
failed with "Are you trying to mount a directory onto a file".
`Dockerfile.graphhopper` copies the file in instead.

**The graph cache must sit in Coolify's managed directory.** The same rewriting
applies to the cache mount: an absolute path and a `GRAPH_CACHE_PATH` variable
were both ignored. Place the cache at
`/data/coolify/applications/<uuid>/graph-cache` on the server, which is what the
relative path in `docker-compose.yml` maps to. Inside the container it must be
`/data/default-gh`, because the image overrides `graph.location` and the value
in `config.yml` is ignored - mount it anywhere else and GraphHopper re-imports
from a `.pbf` that is deliberately not on the server.

**The graphhopper service declares no ports and no `expose`.** Containers on the
same compose network reach each other regardless, so Caddy still proxies to
8989 - but Coolify treats a service with declared ports as web-facing, generates
a domain for it, and Traefik then routes straight to GraphHopper, bypassing the
token gate entirely.

The base image is pinned to a digest in `Dockerfile.graphhopper`. A different
GraphHopper release can reject the existing cache and fall back to an import
that cannot complete.

### Health check

Point Coolify's health check at Caddy's `/health` on port 8080. Both services
also declare their own healthcheck in `docker-compose.yml`.

## scripts/

`extract_waterways.sh` pulls the waterway network out of an OSM extract. It
feeds the motion-planner's dormant waterways profiles, not GraphHopper.
