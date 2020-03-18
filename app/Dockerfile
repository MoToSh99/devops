FROM golang:alpine AS build

RUN apk --no-cache add gcc g++ make git

WORKDIR /devops/src
COPY src .

WORKDIR /devops
COPY go.mod .
COPY go.sum .
COPY wait-for-it.sh .

RUN go mod download

RUN GOOS=linux go build -ldflags="-s -w" -o ./bin/web-app ./src/minitwit.go 

FROM alpine:3.9
RUN apk --no-cache add ca-certificates
WORKDIR /usr/bin

COPY --from=build /devops/src/static /usr/bin/src/static
COPY --from=build /devops/bin /go/bin
COPY --from=build /devops /usr/bin

RUN chmod +x wait-for-it.sh
RUN apk add --update bash

EXPOSE 5000
ENTRYPOINT ["./wait-for-it.sh", "127.0.0.1:5432", "--", "/go/bin/web-app"]
