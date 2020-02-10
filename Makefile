build:
	go build src/minitwit.go
	gcc src/flagtool/flag_tool.c -l sqlite3 -L/opt/local/lib/ -o flag_tool -g

start:
	go run src/minitwit.go

clean:
	rm flag_tool
	rm src/minitwit


inspectdb:
	./flag_tool -i | less

flag:
	./lag_tool "$@"