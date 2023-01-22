FROM golang

WORKDIR /usr/src/app

COPY ./entry/go.mod .
RUN go mod download && go mod verify

COPY ./entry .
COPY ./functions .
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]