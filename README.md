                        / MiniTwit /

    ~ What is MiniTwit?

      A SQLite and Gorilla powered twitter clone written in Golang

    ~ How do I use it?

      Use the Makefile to build and run the solution. Hence:

      make build && make start

      When up and running, the application will greet
      you on http://localhost:5000/

    ~ Is it tested?

      Yes, we have a fully functional automated test suite running.

---

Following commands should be run in the terminal:

To get the newest go version:

`sudo snap install --classic go`

To install all project dependencies:

`make install`

To build and run the project:
`make build && make start`

To add go as environment variable:
export PATH=\$PATH:/usr/local/go/bin

# deploy

The application is hosted on following ip:
178.128.249.71

To get access your public key needs to be uploadet to the server. To login to the server:
`ssh root@178.128.249.71`
Once on the server, the code can be found at:
`~/var/www/devops`
To update the code here, do a git pull. Once pulled deploy by:

```bash
  cd app
  make prod-start
```

---

![Dependency diagram](./dependencydiagram.png)
