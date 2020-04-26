FROM golang:1.14

WORKDIR /go/src/kropotnik
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["kropotnik"]
