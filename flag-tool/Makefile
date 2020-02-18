build:
	gcc src/flagtool/flag_tool.c -l sqlite3 -L/opt/local/lib/ -o bin/flag_tool -g

clean:
	rm ./bin/flag_tool

inspectdb:
	./bin/flag_tool -i | less

flag:
	./bin/flag_tool "$@"