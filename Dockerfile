FROM docker.io/golang:1.26-alpine3.23 as builder

WORKDIR /builder
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build --ldflags="-w -s -X main.addrFormat=:%d" -o main .

FROM docker.io/alpine:3.23

WORKDIR /opt/media
COPY --from=builder /builder/main server
COPY --from=builder /builder/public public
RUN chown nobody:nobody server

USER nobody
ENTRYPOINT ["/opt/media/server"]
