FROM golang:1-alpine
MAINTAINER Jeffrey Jenner <thetooth@ameoto.com>
WORKDIR /go/src/github.com/thetooth/thetooth.name/

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh ca-certificates
RUN go get -u github.com/golang/dep/cmd/dep
COPY . .
RUN /go/bin/dep ensure -v
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

FROM progrium/busybox
MAINTAINER Jeffrey Jenner <thetooth@ameoto.com>
WORKDIR "/opt/thetooth.name"

COPY --from=0 /go/src/github.com/thetooth/thetooth.name/static /opt/thetooth.name/static
COPY --from=0 /go/src/github.com/thetooth/thetooth.name/template.html /opt/thetooth.name/template.html
COPY --from=0 /go/src/github.com/thetooth/thetooth.name/server /opt/thetooth.name/server
RUN chmod 755 /opt/thetooth.name/server

CMD ["./server"]
