FROM golang:latest

RUN go version
ENV GOPATH=/

COPY ./ ./

RUN go build -v ./cmd/apiserver

CMD ["./apiserver"]

