install:
	go install ./src/minitwit.go


#No need to build minitwit...
build:
	go build -o bin/minitwit src/minitwit.go

start:
	go run ./src/minitwit.go

clean:
	rm ./bin/minitwit


prod-start:
	docker-compose up -d --no-deps --build
