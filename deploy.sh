cd $(go env GOPATH)/src/github.com/mattpaletta/AbilitySoftwareGroup468 && git pull origin master
go install github.com/mattpaletta/AbilitySoftwareGroup468
docker image build --file Dockerfile --tag ability:latest .
docker image save ability:latest -o dockerimage
