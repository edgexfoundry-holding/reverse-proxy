FROM golang:1.9-alpine AS builder

RUN mkdir -p /edgexsecurity

WORKDIR /edgexsecurity

COPY . .

RUN apk update && apk upgrade && apk add --no-cache  git

RUN go get github.com/dghubble/sling && go get github.com/BurntSushi/toml && go get github.com/edgexfoundry/edgex-go/support/logging-client && go get github.com/dgrijalva/jwt-go

RUN cd core && go build -o edgexproxy

COPY Docker/res/configuration.toml core/res/

WORKDIR core

ENTRYPOINT ["./edgexproxy"]
CMD  ["--init=true"]