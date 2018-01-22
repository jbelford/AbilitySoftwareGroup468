FROM golang

ADD . /go/src/github.com/mattpaletta/AbilitySoftwareGroup468

RUN go get github.com/mattpaletta/AbilitySoftwareGroup468
RUN go install github.com/mattpaletta/AbilitySoftwareGroup468

EXPOSE 8080
