package main

import (
	"context"
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
		klog.Fatal(err)
	}

	s := http.NewServer()
	s.JournalService = sqlite.NewJournalService(db)
	s.NowService = sqlite.NewNowService(db)
	s.UserService = sqlite.NewUserService(db)

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
	}()

	klog.Infof("Starting the HTTP server")
	err = s.Open("127.0.0.1:1111")
	if err != nil {
		klog.Fatal(err)
	}

	// Wait for CTRL-C
	<-ctx.Done()
}
