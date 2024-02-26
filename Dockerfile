FROM golang:latest

RUN go version
ENV GOPATH=/

COPY ./ ./

RUN go mod download
RUN go build -o sc-proxy ./cmd/proxy/main.go

CMD ["./sc-proxy"]