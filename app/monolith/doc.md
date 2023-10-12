`monolithic` apporach choose for the mvp , it will deploy as single app binary and the application launch as default configuration to crawling the link for every 1 minute and ranking the page for every 1 hour with single database for storing and indexing.


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
