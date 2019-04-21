# DocShelf
[![build](https://gitlab.com/docshelf/docshelf/badges/master/build.svg?job=test)](https://gitlab.com/docshelf/docshelf/pipelines) [![coverage](https://gitlab.com/docshelf/docshelf/badges/master/coverage.svg?job=test)](https://docshelf.gitlab.io/docshelf)
[![documentation](https://godoc.org/github.com/docshelf/docshelf?status.svg)](http://godoc.org/github.com/docshelf/docshelf)

A lightweight, team documentation solution that won't make you pull your hair out.

## !WIP!
This project is still a pre-alpha work in progress and isn't suitable for any real use cases yet. Come back soon though! :smile:

## Quickstart
The fastest way to get up and running with docshelf is to build and run the docker container.
```
$ docker build -t docshelf .
$ docker run -it -p 1337:1337 docshelf
```
Navigating to [http://localhost:1337/](http://localhost:1337/) should pop up a login window.

## Getting Started
To get the docshelf API running natively on your local machine, you just need to have the go compiler installed.
```
$ go run cmd/server/main.go
```
When docshelf starts up, it will spin up a local bolt database, a bleve search index, and all documents will be stored locally in a `documents/` folder.

### AWS
If you want to test docshelf with the AWS backends, all you have to do is set some environment variables. This assumes that your AWS credentials are already present in your environment.

#### Dynamo Backend
```
$ DS_BACKEND=dynamo go run cmd/server/main.go
```
Docshelf will automatically provision the necessary dynamo tables if they don't exist, so give it a minute or two on the first startup.


#### S3 File Store
```
$ DS_FILE_BACKEND=s3 DS_S3_BUCKET=docshelf-test go run cmd/server/main.go
```

## Configuration
Currently, docshelf can only be configured through environment variables. This table shows all of the current options that can be set.

| Var             | Possible Values  | Description                                     |
|-----------------|------------------|-------------------------------------------------|
| DS_BACKEND      | bolt, dynamo     | Backend for users, doc metadata, etc.           |
| DS_FILE_BACKEND | disk, s3         | How to store document content                   |
| DS_TEXT_INDEX   | bleve, elastic\* | What text index to use for search               |
| DS_S3_BUCKET    | string           | The bucket to use with the s3 file backend      |
| DS_FILE_PREFIX  | string           | The path/prefix to apply to all saved documents |
| DS_HOST         | string           | The host for the API to listen on               |
| DS_PORT         | 0-65535          | The port for the API to listen on               |

_\*elastic is not currenlty supported, but will be in the near future_

More configuration options will become available as dochself becomes more full-featured.

## Experimental UI
There's curently a bare bones UI as a nicer way of testing docshelf features than running dozens of postman requests. It's written in Elm, but the front end tech for docshelf hasn't been decided yet, so there's no guarantee this will stick around long term.



#### Install Elm
[Official Instructions](https://guide.elm-lang.org/install.html)

Quick install with NPM:
```
$ npm install -g elm
```


#### Install and Run Parcel
```
$ npm install -g parcel-bundler
$ cd ui/
$ parcel watch index.html
```

The API is configured to serve the UI as well, so you can reach the UI at `http://localhost:1337` for now.

*Note:* You'll need to create a user directly through the API in order to use the UI.
