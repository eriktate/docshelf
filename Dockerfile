# build the API
FROM golang:1.15.2-alpine as build_api

WORKDIR /build

COPY . /build

RUN apk add --no-cache git
RUN CGO_ENABLED=0 GOOS=linux go build -o docshelf cmd/server/main.go


# build the UI
FROM node:lts as build_ui

WORKDIR /build

COPY /build ./ui

RUN rm -rf public/build/
RUN npm install

RUN npm run build


# bring it all together
FROM caddy:2.1.1-alpine

WORKDIR /opt/app
COPY ./Caddyfile ./Caddyfile

# listening on broadcast so that docshelf is easily available from the host machine
ENV DS_HOST=0.0.0.0
COPY --from=build_api /build/docshelf /opt/app/docshelf
COPY --from=build_ui /build/dist /opt/app/ui/dist

CMD ./docshelf
