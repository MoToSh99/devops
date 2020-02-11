build:
	go build src/minitwit.go
	gcc src/flagtool/flag_tool.c -l sqlite3 -L/opt/local/lib/ -o src/flagtool/flag_tool -g

start:
	go run src/minitwit.go

clean:
	rm src/flag_tool
	rm src/minitwit


inspectdb:
	./src/flag_tool -i | less

flag:
	./src/flag_tool "$@"