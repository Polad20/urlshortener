FROM golang:1.24.2-alpine

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN go mod download

RUN go build -o urlshortener -v ./cmd/

EXPOSE 8080

CMD [ "./urlshortener" ]