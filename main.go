package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/table"
	"log"
	"os"
	"todoapp-gin/router"
	"todoapp-gin/todo"

	"github.com/NathanBak/easy-cass-go/pkg/easycass"
)

func main() {
	// Load ENV
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	secConnBundlePath := os.Getenv("SEC_CONN_BUNDLE_PATH")

	// Setup CQL Session
	session, err := easycass.GetSession(clientID, clientSecret, secConnBundlePath)
	if err != nil {
		log.Fatal(err)
	}
	// Check AstraDB Connection
	keyspace := easycass.GetKeyspace(session)
	if keyspace == "" {
		log.Fatal(fmt.Errorf("AstraDB connection failed: no keyspaces found"))
	} else {
		log.Printf("AstraDB connected, working in keyspace: %v", keyspace)
	}
	// Augment CQL Session to CQLX
	xSession, err := gocqlx.WrapSession(session, err)
	if err != nil {
		log.Fatal(err)
	}
	defer xSession.Close()

	router := router.SetupRouter(todo.AstraDB{Table: table.New(todo.AstraTableTodoItems), Session: xSession})
	router.Run(":8080")
}
