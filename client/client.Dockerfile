FROM golang

WORKDIR /usr/src/app

COPY . .

CMD ["go", "run", "main.go"]