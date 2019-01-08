FROM golang:1.8 as builder

RUN go get -u github.com/gin-gonic/gin

WORKDIR /go/src/app
COPY worker.go .
RUN go get github.com/streadway/amqp && /usr/local/go/bin/go run app .

# Application image.
FROM golang:1.8

COPY --from=builder /go/src/app/app /usr/local/bin/app

CMD ["/usr/local/bin/app"]