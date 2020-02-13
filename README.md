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

Følgende commands køres i terminalen;

For at hente nyeste version:

```sudo snap install --classic go```


For at compile og køre projektet bruges Makefile'en.
Dette installerer også alle nødvendige pakker:
```make build && make start```


Tilføj til miljøvariabler:
export PATH=$PATH:/usr/local/go/bin
