## Go sample application
### search engine that can input or search link and ranking the page in GO 



In the nutshell user will submit link (url) from `frontend` and it will insert into db, and backend service will do pipeline process to crawl the link, extract the text and title content for indexing or rangking.Beside `frontend` there is 2 core package `crawler` and `calculator`.

Project structure:
```
root
├── app
│   ├── microservice
│   │   └── ...
│   └── monolith
│       ├── service
│       │   ├── linkcrawler
│       │   ├── pagerank
│       │   └── supervisor.go
│       └── main.go
├── pagerank  (core)
│   └── ...
├── crawler     (core)
│   └── ...
├── frontend
│   └── ...
├── internal
│   └── ...
├── linkgraph
│   ├── graph
│   │    └── graph.go
│   └── store
│       ├── memory
│       └── postgredb
├── partition
│   └── ...
├── textIndex
│   ├── index
│   ├── store
├── main.go
│
├── ...

```

`app` directory is consisted application specific code that how app will deploy either monolith or microservice, for now it will deploy as monolithic approach.

`frontend` is rest-api for communicated with backend , there is also pre-rendering html page (template) for every endpoint.

`internal` consisted of package that have non-application-specific implementation code like httpclient, uuid etc. so it can be reusable and changeable without to modified the core module or logic.

<!-- `services` wire component (app-specific) like crawler and pagerank into link-database or index-database -->

how to build to test , visit the app/... 

