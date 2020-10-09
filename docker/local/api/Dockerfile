FROM golang:1.15.2-alpine

EXPOSE 1337

COPY ./docker/local/api /opt/watcher

WORKDIR /opt/watcher
RUN apk add --no-cache git
RUN CGO_ENABLED=0 GOOS=linux go build -o watcher /opt/watcher/watcher.go \
	&& mkdir -p /opt/app \
	&& mv watcher /opt/app/

WORKDIR /opt/docshelf

ENTRYPOINT ["/opt/app/watcher"]
CMD ["cmd/server/main.go"]
