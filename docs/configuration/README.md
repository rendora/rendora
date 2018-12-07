
# Configuration

## Overview
Rendora has a flexible configuration system, you are free to use YAML, TOML or JSON. Rendora expects the config file to be named `config.yaml`, `config.json`, `config.toml` or `config.yml` and placed in `/etc/rendora/` directory or in Rendora's working directory. Also you can use a custom config file by running `rendora --config /path/to/my/cusom_config.yaml`

Also note that almost all config variables are optional. What is required currently is the backend and frontend urls as defined in `backend.url` and `target.url` respectively (see [examples](#examples) below).

## Details

* `listen` *(optional)* Rendora's proxy listen address and port
    * `address` *(optional)*
        - default value: `0.0.0.0`
    * `port`
        * default value: `3001`
* `cache` *(optional)*
    * `type` *(optional)* Set the type of cache store, it can be currently either `local` which is a cache store embedded in Rendora, `redis` which is Redis of course or you can also disable caching by setting this to `none`
        - allowed values: `local`, `redis` or `none`
        - default: `local`
    -  `timeout` *(optional)* the default timeout in **seconds** for caching
        -  default: `3600` (i.e. 1 hour)
    -  `redis` *(optional)* you may need to configure this only if you set `cache.type` to `redis`
        -  `address` *(optional)*
            -  default: `localhost:6379`
        -  `password` *(optional)*
        -  `db` *(optional)* Redis database number
            -  default value: `0`
        - `keyPrefix` key prefix to make sure there isn't any conflict between Rendora and other applications using Redis
            - default: `__:::rendora:`
- `target`
    - `url` **(required)**, This is the base URL used by the headless Chrome instance controlled by Rendora to request pages corresponding to whitelisted requests. You can simply set it to your website url  e.g. `https://example.com`). However, for mainly performance reasons, you can set it to an internal address depending on your architecture while making sure that headless Chrome can address your webapp javascript files necessary to do SSR which is Rendora's goal in the first place. As a hint you may have one of these architectures:

        * **webapp js file/files is/are hosted by the backend server**: then you can set the address to the backend server address
        * **webapp js file/files is/are hosted by the frontend server (e.g. nginx)**: then you can set the address to the frontend server address
        * **webapp js file/files is/are hosted by a CDN or an external server**: then you can set the address to the backend server address

- `backend`
    - `url` **(required)**, the base url of the backend server
- `headless` *(optional)*, this contains the config related to the headless Chrome instance controlled by Rendora
    - `waitAfterDOMLoad` *(optional)*, timeout in milliseconds to wait after the initial DOM load event, you may only what to use it for async apps where you start fetching content after the intial load
        - default: `0`
    - `internal`
        - `url` *(optional)*, this is the address of the headless Chrome instance
            - default: `http://localhost:9222`
    - `blockedURLs` *(optional)*, the headless Chrome normally fetches all requests while rendering the HTML,  that includes all CSS, jpg, gif, analytics js and any other unnecessary asset; some experiments on complex pages have shown a reduction by more than 50% just by blocking all urls except for just the webapp javascript files which are of course necessary to render the page correctly in the first place. You're only allowed to use full urls or wildcards.
        - default: `["*.png", "*.jpg", "*.jpeg", "*.webp", "*.gif", "*.css", "*.woff2", "*.svg", "*.woff", "*.ttf",
		"https://www.youtube.com/*", "https://www.google-analytics.com/*",
		"https://fonts.googleapis.com/*"]`
- `output`
    - `minify` *(optional)*, minify the SSR'ed HTML, this is done before caching so that it doesn't get executed for every whitelisted request
- `filters`
    - `preset` *(optional)*, if the preset is set to `bots` then the user agent is checked against some default *keywords* to see whether the request is whitelisted or not. You can override these keywords in `filters.userAgent.keywords` as shown below.
    - `userAgent`
        - `defaultPolicy` *(optional)*, The default policy of whether the user agents should be whitelisted (i.e. get SSR'ed) or blacklisted (i.e. just return the initial HTML coming from the backend server)
            - allowed values: `whitelist` and `blacklist`
            - default: `whitelist`
        - `keywords` *(optional)*, The allowed keywords (in lowercase since request user agents are converted to lowercase before testing them against keywords) in the request's user agent, if it contains any of these keywrords then the request is considered whitelisted, otherwise it is blacklisted. Keywords are only used when  `filters.preset` is set to `bots` and it's totally ignored when it is set to `all`.
            - default: `["bot", "google", "slurp", "yandex", "crawler", "search"]`
        - `exceptions` *(optional)* You can also add exceptions against the default policy, if `defaultPolicy` is set to `whitelist`, then exceptions are blacklisted and vice versa.
            - example: `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.67 Safari/537.36`
        - `paths` *(optional)*, Paths are checked only if the request user agent is checked and passes its filters
            - `defaultPolicy` *(optional)*, if the default policy is "whitelist" then any path is whitelisted, if it is "blacklist" then all paths are blacklisted
                - allowed values: `whitelist` and `blacklist`
                - default: `whitelist`
            - `exceptions` *(optional)*, Exceptions are the blacklisted paths if the the default policy is `whitelist` and vice versa, there are 2 types of exceptions, if you want to add /posts/*, you can simply add `/posts/` as a prefix
                - `prefix`
                    - default: empty list
                - `exact`
                    - default: empty list
- `debug`: *(optional)*, you usually need to set this to `default` in production
    - default: `false`
- `server`: *(optional)*, contains configuration about Rendora's API server [read more about Rendora's API](/docs/api/)
    - `enable`: *(optional)*
        - default: `false`
    - `listen`: *(optional)*
        - `address`: *(optional)*, listen address if enabled
            - default: `0.0.0.0`
        - `port`: *(optional)*, listen port if enabled
            - default: `9242`
    - `auth`: *(optional)*, optionally set an authentication header name and value 
        - `enable`: *(optional)*
            - default: `false`
        - `name`: *(optional)*, the HTTP authentication header name if enabled
            - default: `X-Auth-Rendora`
        - `value`: *(optional)*, the HTTP authentication header value if enabled, **it's your responsibility to generate a securely random token.**

## Examples

### A minimal config file
```yaml
target:
    url: "http://127.0.0.1" 
backend:
    url: "http://127.0.0.1:8000"
```

### A more customized config file

```yaml
listen:
    address: 0.0.0.0
    port: 3001
cache:
    type: redis
    timeout: 6000
    redis:
        address: localhost:6379
target:
    url: "http://127.0.0.1" 
backend:
    url: "http://127.0.0.1:8000"
headless:
    waitAfterDOMLoad: 0
    internal:
      url: http://localhost:9222
output:
    minify: true
filters:
    preset: bots
    userAgent:
        defaultPolicy: whitelist
        keywords:
           - bot
           - slurp
           - crawler
        exceptions:
            - Mozilla/5.0 (compatible; AhrefsBot/5.2; +http://ahrefs.com/robot/)
    paths:
        defaultPolicy: whitelist
        exceptions:
            prefix:
             - /posts/
             - /users/
            exact:
             - /
             - /about
             - /faq
```