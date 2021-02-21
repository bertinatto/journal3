package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/bertinatto/journal3/http"
	"github.com/bertinatto/journal3/sqlite"
)

func main() {
	db := sqlite.NewDB("blog.db")
	err := db.Open()
	if err != nil {
		log.Fatal(err)
	}

	journalService := sqlite.NewJournalService(db)

	s := http.NewServer()
	s.JournalService = journalService

	// Setup signal handlers.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	err = s.Open()
	if err != nil {
		log.Fatal(err)
	}

	// Wait for CTRL-C.
	<-ctx.Done()

}
