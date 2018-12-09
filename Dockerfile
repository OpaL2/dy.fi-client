FROM golang:1.10

WORKDIR /go/src/github.com/OpaL2/dy.fi-client/
COPY . .

ENV CGO_ENABLED=0 GOOS=linux
RUN go build -a -installsuffix cgo  -o main

FROM scratch
ENV PASSWORD=pass USERNAME=user HOSTNAME=host.fi
COPY --from=0 /go/src/github.com/OpaL2/dy.fi-client/main /

CMD ["/main"]