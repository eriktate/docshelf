FROM caddy:2.1.1-alpine

ENV CADDY_ADDRESS=0.0.0.0
ENV API_ADDRESS=docshelf-api:1337
ENV UI_ADDRESS=docshelf-ui:5000

WORKDIR /opt/caddy
COPY ./Caddyfile ./Caddyfile

ENTRYPOINT ["caddy"]
CMD ["run"]
