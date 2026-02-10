#!/bin/bash
set -e

INPUT="../maps/bene.osm.pbf"
TMP="waterways.osm.pbf"
OUTPUT="waterways.geojson"

osmium tags-filter "$INPUT" w/waterway -o "$TMP" --overwrite
osmium export "$TMP" -f geojson -o "$OUTPUT" --overwrite --geometry-types=linestring
