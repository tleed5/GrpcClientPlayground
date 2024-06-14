# syntax=docker/dockerfile:1
FROM golang:1.22.4 as build-prod

WORKDIR /build
COPY go.mod go.sum ./

# cache the go mod download where possible
RUN go mod graph | awk '{if ($1 !~ "@") print $2}' | xargs go get

COPY . .
# CGO_ENABLED=0 creates a statically linked binary
# -ldflags "-s -w" remove debug symbols
RUN env CGO_ENABLED=0 go build -ldflags "-s -w" -o "/bin/GrpcClientPlayground"
COPY *.pem /bin/

FROM scratch as production
COPY --from=build-prod /bin/GrpcClientPlayground /bin/GrpcClientPlayground
COPY --from=build-prod /bin/client-cert.pem /bin/
COPY --from=build-prod /bin/client-key.pem /bin/
ENTRYPOINT ["/bin/GrpcClientPlayground"]
