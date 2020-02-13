install:
	go install ./src/minitwit.go


#No need to build minitwit...
build:
	go build -o bin/minitwit src/minitwit.go
	gcc src/flagtool/flag_tool.c -l sqlite3 -L/opt/local/lib/ -o bin/flag_tool -g


start:
	go run ./src/minitwit.go

clean:
	rm ./bin/flag_tool
	rm ./bin/minitwit


inspectdb:
	./bin/flag_tool -i | less

flag:
	./bin/flag_tool "$@"