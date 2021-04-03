package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/bertinatto/journal3/http"
	"github.com/bertinatto/journal3/sqlite"
	"k8s.io/klog"
)

const (
	defaultDataFile = "data.db"
	defaultAddress  = "localhost:8080"
)

func main() {
	file := flag.String("file", defaultDataFile, "file where data will persist")
	domain := flag.String("domain", "", "domain")
	addr := flag.String("listen", defaultAddress, "ip:port")
	flag.Parse()

	dir := filepath.Dir(*file)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		log.Printf("isnotexist")
		err := os.Mkdir(dir, 750)
		if err != nil {
			klog.Fatal(err)
		}

		log.Printf("created %v", dir)
	}
	if err != nil {
		klog.Fatal(err)
	}

	db := sqlite.NewDB(*file)
	err = db.Open()
	if err != nil {
		klog.Fatal(err)
	}

	s := http.NewServer()
	s.Domain = *domain
	s.Addr = *addr
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
	err = s.Open()
	if err != nil {
		klog.Fatal(err)
	}

	if s.TLS() {
		go func() {
			klog.Fatal(http.ListenAndServerTLSRedirect(s.Domain))
		}()
	}

	// Wait for CTRL-C
	<-ctx.Done()
}
