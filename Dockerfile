FROM golang:1.19-alpine AS builder

ADD . /opt/memo

WORKDIR /opt/memo

RUN apk --update add --no-cache ca-certificates openssl git tzdata && \
  update-ca-certificates

RUN go get -v && \
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o memod cmd/memod/* && \
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o memo-cli cmd/memo-cli/* && \
  chmod +x memod memo-cli

FROM alpine:latest

COPY --from=builder /opt/memo/memod /bin/memod
COPY --from=builder /opt/memo/memo-cli /bin/memo-cli
ADD generate-config-file.sh /bin/generate-config-file
RUN chmod +x /bin/generate-config-file

CMD [ "memod" ]