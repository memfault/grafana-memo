FROM golang:1.19-alpine AS builder

ADD . /opt/memo

WORKDIR /opt/memo

RUN apk --update add --no-cache ca-certificates openssl git tzdata && \
  update-ca-certificates

RUN go get -v && \
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o memod cmd/memod/main.go && \
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o memo-cli cmd/memo-cli/main.go && \
  chmod +x memod memo-cli

FROM alpine:latest

COPY --from=builder /opt/memod /bin/memod
COPY --from=builder /opt/memo-cli /bin/memo-cli

CMD [ "memod" ]