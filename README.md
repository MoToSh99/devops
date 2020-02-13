                        / MiniTwit /

    ~ What is MiniTwit?

      A SQLite and Gorilla powered twitter clone written in Golang

    ~ How do I use it?
    
      Use the Makefile to build and run the solution. Hence:
      
      make build && make start

      When up and running, the application will greet
      you on http://localhost:5000/
	
    ~ Is it tested?

      At the moment the solution is only tested manually. 
      Hence no automated testing.

----------------------------------------------------------------

Following commands should be run in the terminal:

To get the newest go version:

```sudo snap install --classic go```


To install all project dependencies:

```make install```

To build and run the project:
```make build && make start```


To add go as environment variable:
export PATH=$PATH:/usr/local/go/bin
