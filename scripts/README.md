# Route Experiments

This experiment reads coordinates from `coordinates.json` and computes car routes using GraphHopper.

## Prerequisites

1. GraphHopper API must be running (via docker-compose)
2. Go must be installed on your system

## Input Format

The `coordinates.json` file should contain an array of route requests:

```json
[
  {
    "id": "route_1",
    "from": {
      "lat": 52.3676,
      "lon": 4.9041
    },
    "to": {
      "lat": 52.3702,
      "lon": 4.8952
    }
  }
]
```

## Running the Experiment

1. Make sure GraphHopper is running:
   ```bash
   docker-compose up -d
   ```

2. Navigate to the scripts directory:
   ```bash
   cd scripts
   ```

3. Run the experiment:
   ```bash
   go run experiments.go
   ```

## Output

The program generates `route_results.json` containing:
- Original coordinates (from/to)
- Distance in meters
- Time in seconds
- Full GraphHopper API response
- Any errors encountered

## Configuration

You can modify these constants in `experiments.go`:
- `graphhopperURL`: GraphHopper API endpoint (default: http://localhost:8989/route)
- `profile`: Routing profile to use (default: car)
