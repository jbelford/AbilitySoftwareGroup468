FROM golang:1.9.2 as builder
ADD . /go/src/github.com/mattpaletta/AbilitySoftwareGroup468
WORKDIR /go/src/github.com/mattpaletta/AbilitySoftwareGroup468
RUN go get ./...
RUN go install github.com/mattpaletta/AbilitySoftwareGroup468
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -tags netgo -installsuffix netgo -o app .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -tags netgo -installsuffix netgo -o workloadgen workloads/workloadgenerator.go

#FROM alpine:3.7 as certs
#ENV PATH=/bin
#RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

FROM alpine:latest
ENV PATH=/bin
#COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR /root/
ADD workloads workloads
ADD config config
COPY --from=builder /go/src/github.com/mattpaletta/AbilitySoftwareGroup468/app .
COPY --from=builder /go/src/github.com/mattpaletta/AbilitySoftwareGroup468/workloadgen .
ENTRYPOINT ["./app"]
