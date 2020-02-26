build:
	docker build -t flag-tool .

clean:
	docker rmi flag-tool

inspectdb:
	docker run -it --rm flag-tool -i | less

#to flag messages, use shell script