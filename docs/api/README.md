
# API

Rendora can be configured when the config `server.enable` is set to `true` to provide another HTTP server listening to the port `9242` by default (can be changed using the config file) in order to provide more info and metrics. Currently there are 2 HTTP endpoints

* **rendering**: provides a JSON response that contains the SSR'ed HTML page, its status code and headers.
    * endpoint: `POST /render`
    * request body: A serialized json object that contains:
        * `uri`: the request uri (e.g. `/posts`)
    * response body: A serialized json object that contains:
        * `body`: the SSR'ed HTML page
        * `status`: the status code
        * `headers`: response headers
* **metrics**: provides Prometheus metrics
    * endpoint: `GET /metrics`
    * Rendora's metrics:
        * `rendora_requests_total`: provides a counter corresponding to the number of total requests (i.e. both whitelisted and blacklisted requests)
        * `rendora_requests_ssr`: provides a counter corresponding to the number of total whitelisted requests
        * `rendora_requests_ssr_cached`: provides a counter corresponding to the number of cached whitelisted requests
        * `rendora_latency_ssr`: provides a historgram for SSR latency in milliseconds for uncached SSR'ed requests with buckets of values `[50, 100, 150, 200, 250, 300, 350, 400, 500]`