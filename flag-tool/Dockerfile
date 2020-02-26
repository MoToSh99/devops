FROM golang:alpine AS build

RUN apk --no-cache add gcc g++ make git

WORKDIR /flagTool/src
COPY src .


WORKDIR /flagTool
COPY go.mod .
COPY go.sum .

RUN go mod download

RUN GOOS=linux go build -ldflags="-s -w" -o ./bin/flagTool ./src/flagTool.go 

FROM alpine:3.9
RUN apk --no-cache add ca-certificates
WORKDIR /usr/bin

COPY --from=build /flagTool/bin /go/bin

ENTRYPOINT ["/go/bin/flagTool"]
