FROM golang

ADD . /go/src/github.com/mattpaletta/AbilitySoftwareGroup468
RUN go get ./...
RUN go install github.com/mattpaletta/AbilitySoftwareGroup468
WORKDIR /go/src/github.com/mattpaletta/AbilitySoftwareGroup468
