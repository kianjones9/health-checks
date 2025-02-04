# health-checks

## Quickstart

### Run in a Docker Container (Recommended)
Using default monitors.yaml
```
docker run kianjones9/health-checks
```
Or mounting your custom monitors.yaml file
```
docker run -v ./monitors.yaml:/app/monitors.yaml kianjones9/health-checks
```
Or build it from source
```
docker build -t namespace/image-name:tag-name .
```

### Build from source
Dependencies:
- git
- go
- make

```
git clone git@github.com:kianjones9/health-checks.git
cd health-checks
make build
make run
```


## Context

This is my attempt at the take home coding assessment given to me by [Fetch](https://www.fetch.com) as part of my interview process for the position of Site Reliability Engineer.

The problem statement is as follows:

"Implement a program to check the health of a set of HTTP endpoints... Read an input argument to a file path with a list of HTTP endpoints in YAML format. Test the health of the endpoints every 15 seconds. Keep track of the availability percentage of the HTTP domain names being monitored by the program. Log the cumulative availability percentage for each domain to the console after the completion of each 15-second test cycle... For the purposes of this exercise, it is fine to use a suitable data structure in your application’s memory to keep track of and log the expected program output over time as each testing cycle completes."

## Design Considerations
There are a few key considerations I made I'd like to highlight.

1. Based on the spec set out by Fetch, an endpoint is considered "up" if and only if the response's status code is `2xx` (any 200–299 response code) and the response latency is less than 500 ms. In practice, we may consider status codes beyond just this range as successful, for example a 301 redirect may be the anticipated status code of a particular endpoint, and thus not contribute to a poor availability score.

2. Additionally, accurately measuring latency can be tricky. My solution implements Go's `ClientTrace` from the `net/http/httptrace` package, which allows us to specify an anonymous callback function on the `GotFirstResponseByte` trigger, and arrive at a pretty accurate timing. However, in my testing, I determined that even with this technique, we still have a ~1ms.

3. The problem state also calls for binning the monitors by domain, such that endpoints poiting to different resources (i.e. different URIs) have their availability reported by domain. I chose to implement the probing in a way which allows us to tabulate availability by URI or even by page in future.