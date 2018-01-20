# Build Image
FROM golang:1.9-alpine as build
RUN apk update && apk add ca-certificates
RUN apk add --no-cache git
RUN go get github.com/golang/dep/cmd/dep
WORKDIR /go/src/
RUN git clone https://github.com/robertcsapo/cisco-talos-tcpwrapper
WORKDIR /go/src/cisco-talos-tcpwrapper/
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Create runtime
FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/src/cisco-talos-tcpwrapper/main /go/src/cisco-talos-tcpwrapper/main
WORKDIR /go/src/cisco-talos-tcpwrapper/
ENTRYPOINT ["./main"]
