FROM golang

WORKDIR /usr/src/app

COPY ./client .
COPY ./functions/aes.go ./aes.go

CMD ["go", "run", "main.go", "aes.go"]