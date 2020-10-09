FROM node:lts as build_ui

EXPOSE 5000
WORKDIR /opt/ui

COPY ./ui /opt/ui

# clear any existing build artifacts
RUN rm -rf public/build/
RUN rm -rf node_modules/

# install deps
RUN npm install

# build bundle
RUN npm run build

ENTRYPOINT ["npm"]
CMD ["start", "--", "--host"]
