FROM golang:1.7

RUN mkdir -p /app

WORKDIR /app

ADD . /app

RUN go build ./netspresso.go ./client.go ./synthetic.go

CMD ["./netspresso", "-a", "sun", "-c", "1"]
