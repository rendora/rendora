---
title: FAQs
---
# FAQs
## What is Dynamic Rendering?
Dynamic rendering means that the server provides server-side rendered HTML to web crawlers such as GoogleBot and BingBot and at the same time provides the typical initial HTML to browsers and real users in order to be rendered in client-side. Dynamic rendering is used to improve SEO for websites written in modern javascript frameworks such as React.js, Vue.js and Angular.js.


## What is the difference between Rendora and Puppeteer?

[Puppeteer](https://github.com/GoogleChrome/puppeteer) is a great Node.js library which provides a generic high-level API to control Chrome. On the other hand, Rendora is a dynamic renderer that acts as a reverse HTTP proxy placed in front of your backend server to provide server-side rendering mainly to web crawlers in order to effortlessly improve SEO.


## What is the difference between Rendora and Rendertron?
[Rendertron](https://github.com/GoogleChrome/rendertron) is comparable to Rendora in the sense that they both aim to provide SSR using headless Chrome; however there are various differences that can make Rendora a much better choice:


1. **Architecture**: Rendertron is a HTTP server that returns SSR'ed HTML back to the client. That means that your server must contain the necessary code to filter requests and asks rendertron to provide the SSR'ed HTML and then return it back to the original client. Rendora does all that automatically by acting as a reverse HTTP proxy in front of your backend.


2. **Caching**: Rendora can be configured to use internal local store or Redis to cache SSR'ed HTML.
3. **Performance**: In addition to caching, Rendora is able to skip fetching and rendering unnecessary content CSS, fonts, images, etc... which can substantially reduce the intial DOM load latency.
4. **Development**: Rendertron is developed in Node.js while Rendora is a single binary written in Golang.
5. **API and Metrics**: Rendora provides Prometheus metrics about SSR latencies and number of SSR'ed and total requests. Furthermore, Rendora provides a JSON rendering endpoint that contains body, status and headers of the SSR response by the headless Chrome instance.