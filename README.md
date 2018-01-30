# AbilitySoftwareGroup468

[![Build Status](https://travis-ci.com/mattpaletta/AbilitySoftwareGroup468.svg?token=ysncAybhRTtbpjrpSW8S&branch=master)](https://travis-ci.com/mattpaletta/AbilitySoftwareGroup468)

## To test on your local machine (in docker).
`
docker-compose up -d
`

## Optionally, if `Dockerfile` or `dockercompose.yml` have changed,
`
docker-compose up -d --build
`

### To test with larger scale systems (Eg. 10 instances of workloadgen):
`
docker-compose up -d --scale workloadgen:10
`



## To view running images:
`
docker-compose ps
`

## And to stop
`
docker-compose stop
`
## To stop, and remove:
`
docker-compose down
`

### Available Ports
`
44420-44424
`

## Current use of ports/hostnames:
* Mongo: `db1.prod.ability.com:44423`
* Transaction Server: `transaction.prod.ability.com:44421`
* Web Server: `webserver.prod.ability.com:44420`
* Audit Server: `auditserver.prod.ability.com:44422`

## How to get into the workload generator running in docker:
Once the services are running in `docker-compose`, which you can verify with `docker-compose ps`,
Use `docker-compose exec workloadgen bash` to start a bash prompt in the workload generator.
Once inside, run `go run workloadgenerator.go <nameoftestfile> <numthreads>`.  All files from `workloads/` are included.
Alternatively: `docker-compose down && docker-compose up -d --build && docker-compose exec workloadgen bash`
To get out of the container `exit`


### To delete docker images/containers
`docker rmi $(docker images -q)`
`docker rm $(docker ps -a -q)`
