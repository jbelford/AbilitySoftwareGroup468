cd $(go env GOPATH)/src/github.com/mattpaletta/AbilitySoftwareGroup468 && git pull origin master
go install github.com/mattpaletta/AbilitySoftwareGroup468
docker build --file Dockerfile --tag ability:latest

