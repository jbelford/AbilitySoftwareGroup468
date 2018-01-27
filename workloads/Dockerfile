#FROM golang:1.9.2 as builderworkload
#ADD . /go/src/github.com/mattpaletta/AbilitySoftwareGroup468
#WORKDIR /go/src/github.com/mattpaletta/AbilitySoftwareGroup468
#RUN go get ./...
#RUN go install github.com/mattpaletta/AbilitySoftwareGroup468
#RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -gcflags="-m64 -pthread -fmessage-length=0 -fdebug-prefix-map=/tmp/go-build075062417=/tmp/go-build -gno-record-gcc-switches" -tags netgo -installsuffix netgo -o workloadgen workloadgenerator.go

#FROM alpine:3.7 as certs
#ENV PATH=/bin
#RUN  apk add --update ca-certificates && rm -rf /var/cache/apk/*

FROM golang:1.9.2
COPY . /root/workloads
WORKDIR /root/workloads/
#COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
#COPY --from=builderworkload /go/src/github.com/mattpaletta/AbilitySoftwareGroup468/workloadgen .
#RUN chmod +x ./workloadgen
ENTRYPOINT ["bash"] 
#["./workloadgen"]
