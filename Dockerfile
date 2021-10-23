FROM golang:1.16
ADD . /go/src


RUN cd /go/src/ ;go mod download; go build -o /go/cache-api main.go

ENTRYPOINT ["/go/cache-api"]

# Expose the server TCP port
EXPOSE 8080