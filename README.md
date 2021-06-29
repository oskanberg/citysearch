
# City Search

This Go service provides an API that suggests city names based on a query, and optional latitude/longitude location.

The search results are ranked by their match scores, which are inversely proportional to a combination of:

- string match accuracy, judged by Levenshtein distance from the target name (case insensitive)

- their euclidean distance from the given latitude/longitude, if specified


## Usage

The service can either be built/run directly with Go, or as a docker image.

With Go installed, `go run cmd/citysearch/main.go --cities=cities15000.csv --port=8080`

With Docker, `docker build -t citysearch . && docker run -p 8080:80 citysearch`

Both of these options should start the service on localhost port `8080`.


### Flags

Run directly, the service accepts two flags:

`--cities` is required, and locates the csv database of cities to use.

`--port` optionally specifies the port to serve on. By default this is `:80`.


# Endpoint

`GET /suggestions?q=[&latitude=&longitude=]`

`q` is required, and is the string query for searching

`latitude` and `longitude` are optional, but must both be specified if used. They represent the location of the caller, and is used to modulate the results to be location-specific.

## Example

`GET /suggestions?q=Chi&latitude=50.83673&longitude=-0.78003`

```json
{
    "suggestions": [
        {
            "name": "Chichester",
            "latitude": 50.83673,
            "longitude": -0.78003,
            "score": 0.5714285714285714
        },
        {
            "name": "Christchurch",
            "latitude": 50.73583,
            "longitude": -1.78129,
            "score": 0.06247351444914479
        }
    ]
}
```