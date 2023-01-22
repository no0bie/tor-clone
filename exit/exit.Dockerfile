FROM golang

WORKDIR /usr/src/app

COPY ./exit/go.mod .
RUN go mod download && go mod verify

COPY ./exit .
COPY ./functions .
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]