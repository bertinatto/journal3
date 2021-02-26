package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/bertinatto/journal3/http"
	"github.com/bertinatto/journal3/sqlite"
	"k8s.io/klog"
)

func main() {
	db := sqlite.NewDB("blog.db")
	err := db.Open()
	if err != nil {
		log.Fatal(err)
	}

	journalService := sqlite.NewJournalService(db)
	nowService := sqlite.NewNowService(db)

	s := http.NewServer()
	s.JournalService = journalService
	s.NowService = nowService

	// Setup signal handlers.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; cancel() }()

	klog.Infof("Starting the HTTP server")
	err = s.Open()
	if err != nil {
		log.Fatal(err)
	}

	// Wait for CTRL-C.
	<-ctx.Done()

}
