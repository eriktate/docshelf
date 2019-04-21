# build the API
FROM golang:1.12.4-alpine as build_api

WORKDIR /build

COPY . /build

RUN apk add --no-cache git
RUN CGO_ENABLED=0 GOOS=linux go build -o docshelf cmd/server/main.go


# build the UI
FROM node:lts as build_ui

WORKDIR /build

COPY --from=build_api /build/ui /build

RUN rm -rf ~/.elm
RUN rm -rf elm-stuff/
RUN rm -rf dist/
RUN npm install elm
RUN npm install parcel-bundler
# RUN npm install --unsafe-perm=true --allow-root -g elm
# RUN npm install --unsafe-perm=true --allow-root -g parcel-bundler
# RUN npm install --unsafe-perm=true --allow-root
# have to manually run parcel, because reasons
RUN node ./node_modules/parcel-bundler/bin/cli.js build index.html


# bring it all together
FROM alpine

WORKDIR /opt/app

# listening on broadcast so that docshelf is easily available from the host machine
ENV DS_HOST=0.0.0.0
COPY --from=build_api /build/docshelf /opt/app/docshelf
COPY --from=build_ui /build/dist /opt/app/ui/dist

CMD ./docshelf
