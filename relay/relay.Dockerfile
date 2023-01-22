FROM golang

WORKDIR /usr/src/app

COPY ./relay/go.mod .
RUN go mod download && go mod verify

COPY ./relay .
COPY ./functions .
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]