---
title: Introduction
---

# Introduction
## What is Rendora?

Rendora can be seen as a reverse HTTP proxy server sitting between your backend server (e.g. Node.js/Express.js, Python/Django, etc...)
and potentially your frontend proxy server (e.g. nginx, traefik, apache, etc...) or even directly to the outside world that does actually nothing but transporting requests and responses as they are **except** when it detects whitelisted requests according to the config. In that case, Rendora instructs a headless Chrome instance to request and render the corresponding page and then return the server-side rendered page back to the client (i.e. the frontend proxy server or the outside world). This simple functionality makes Rendora a powerful dynamic renderer
without actually changing anything in both frontend and backend code.

![Diagram](./pics/diagram.png)


## What is Dynamic Rendering?
Dynamic rendering means that the server provides server-side rendered HTML to web crawlers like GoogleBot and BingBot and at the same time provides the typical initial HTML to normal users in order to be rendered in client-side. Dynamic rendering is meant to improve SEO for websites written in modern javascript frameworks like React, Vue, Angular, etc... You might want to read more about dynamic rendering as recommended from
[Google](https://developers.google.com/search/docs/guides/dynamic-rendering) and 
[Bing](https://blogs.bing.com/webmaster/october-2018/bingbot-Series-JavaScript,-Dynamic-Rendering,-and-Cloaking-Oh-My) to understand more about it.


## Main Features
1. **Zero setup**: no change needed no matter what your frontend javascript framework and/or your backend language/framework are.
2. **High performance**: Rendora uses caching (local inside Rendora's binary or Redis) and controls headless Chrome to skips rendering unncessary assets like images, fonts and CSS to speed up the DOM load.
3. **Highly Configurable**: Control whitelisted user agents and paths among many other configurations.
4. **API**: Rendora has an optional API server that provides Prometheus metrics and JSON rendering endpoint.


## How does Rendora work?

For every request coming from the frontend server or the outside world, there are some checks or filters that are tested against the headers and/or paths according to Rendora's configuration file to determine whether Rendora should just pass the initial HTML returned from the backend server or use headless Chrome to provide a server-side rendered HTML. To be more specific, for every request there are 2 paths:

1. If the request is whitelisted as a candidate for SSR (i.e. a GET request that passes all user agent and path filters), Rendora instructs the headless Chrome instance to request the corresponding page, render it and return the response which contains the final server-side rendered HTML. You usually want to whitelist only web crawlers like GoogleBot, BingBot, etc...

2. If the request isn't whitelisted (i.e. the request is not a GET request or doesn't pass any of the filters), Rendora will simply act as a transparent reverse HTTP proxy and just conveys requests and responses as they are. You usually want to blacklist real users in order to return the usual client-side rendered HTML coming from the backend server back to them.





## Install and run Rendora

### First, run a headless Chrome instance
If Chrome/Chromium is installed in your system, you can run it using

``` bash
google-chrome --headless --remote-debugging-port=9222
```
or simply using docker

``` bash
docker run --tmpfs /tmp --net=host rendora/chrome-headless
```

*note:* the `tmpfs` flag is optional but it's recommended for performance reasons since `rendora/chrome-headless` runs with flag `--user-data-dir=/tmp`

### Then, run Rendora

you can build and run Rendora from source code, (**NOTE**: please read the [configuration manual](configuration/) before running Rendora)

``` bash
git clone https://github.com/rendora/rendora
cd rendora
# MAKE SURE YOU HAVE GO V1.11+ INSTALLED
make build
sudo make install
rendora --config CONFIG_FILE.yaml
```

or simply using docker

``` bash
docker run --net=host -v ./CONFIG_FILE.yaml:/etc/rendora/config.yaml rendora/rendora
```


