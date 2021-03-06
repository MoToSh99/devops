package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/matt035343/devops/app/src/database"
	"github.com/matt035343/devops/app/src/utils"
)

const docStr = `ITU-Minitwit Tweet Flagging Tool

Usage:
  flag_tool <tweet_id>...
  flag_tool -i
  flag_tool -h
Options:
  -h            Show this screen.
  -i            Dump all tweets and authors to STDOUT.`

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing argument")
		return
	}

	arg := os.Args[1]
	utils.InitEnvironmentVariables()
	connectionString := "host=127.0.0.1 port=5432 user=" + utils.GetEnvironmentVariable("POSTGRES_USER") + " dbname=" + utils.GetEnvironmentVariable("POSTGRES_DB") + " password=" + utils.GetEnvironmentVariable("POSTGRES_PASSWORD") + " sslmode=disable"
	db, err := database.ConnectDatabase("postgres", connectionString)
	defer db.CloseDatabase()

	if err != nil {
		fmt.Println("Can't open database")
		fmt.Println(err)
		return
	}

	if arg == "-h" {
		fmt.Println(docStr)
		return
	}

	if arg == "-i" {
		messages, err := db.GetAllMessages()
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, message := range messages {
			fmt.Printf("%d,%d,%s,%s,%t\n",
				message.ID,
				message.AuthorID,
				message.Text,
				time.Unix(message.PublishedDate, 0).Format(time.RFC822),
				message.Flagged,
			)
		}
		return
	}

	for i := 1; i < len(os.Args); i++ {
		ID, err := strconv.Atoi(os.Args[i])
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = db.FlagMessage(ID)
		if err != nil {
			fmt.Println("SQL error")
			fmt.Println(err)
			continue
		}
		fmt.Printf("Flagged entry: %d\n", ID)
	}

}
