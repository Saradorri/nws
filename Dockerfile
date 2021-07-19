FROM golang:1.15-alpine
WORKDIR /nws
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -i -o ./build/nws
ENTRYPOINT ["./build/nws"]
