├## Go sample application
### search engine that can input or search link and ranking the page in GO 



In the nutshell user will submit link (url) from `frontend` and it will insert into db, and backend service will do pipeline process to crawl the link, extract the text and title content for indexing or rangking.Beside `frontend` there is 2 core package `crawler` and `calculator`.

Project structure:
```
root
├── app
│   ├── microservice
│   │   └── ...
│   └── monolith
│       ├── doc.go
│       ├── supervisor.go
│       └── main.go
├── frontend
│   └── ...
├── internal
│   └── ...
├── linkcrawler (core service)
│   └── ...
├── linkgraph (domain)
│   └── ...
├── pagerank  (core service)
│   └── ...
├── partition
│   └── ...
|── store
│   └── ...
├── textIndex (domain)
│   └── ...
├── main.go
│
├── ...

```

`app` directory is consisted application specific code that how app will deploy either monolith or microservice, for now it will deploy as monolithic approach.

`frontend` is rest-api for communicated with backend , there is also pre-rendering html page (template) for every endpoint.

`internal` consisted of package that have non-application-specific implementation code like httpclient, uuid etc. so it can be reusable and changeable without to modified the core module or logic.

`link crawler` service background that process the submitted

`linkgraph` manage of link and edge

`pagerank` calculate ranking for indexing

`store` implementation of linkgraph

`textIndex` manage indexed link 

<!-- `services` wire component (app-specific) like crawler and pagerank into link-database or index-database -->

how to build for try , visit the app/... 

