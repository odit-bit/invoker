## Go sample application
### search engine that can input or search link and ranking the page in GO 

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
├── calculator  (core)
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
├── service
│   ├── linkcrawler
│   ├── pagerank
│   └── supervisor.go
├── textIndex
│   ├── index
│   ├── store
└── main.go
...

```

`app` directory is consisted application specific code that how app will deploy either monolith or microservice, for now it will deploy as monolithic approach.

`frontend` is rest-api for communicated with backend , there is also pre-rendering html page (template) for every endpoint.

beside `frontend` there is 2 core package `crawler`,`calculator`.
in the nutshell user will submit link (url) from `frontend` and it will insert into db, and backend will do background process for submited link



`internal` consisted of package that have non-application-specific implementation code like httpclient, pipeline, uuid etc.

<!-- `services` wire component (app-specific) like crawler and pagerank into link-database or index-database -->


using docker at remote host.
in local terminal enter this
```
 export DOCKER_HOST=ssh://username@remote-ip
```
all docker command will connect to remote host as long as terminal is not close.

and run the container 
```
docker compose up
```
