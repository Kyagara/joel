package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/disgoorg/disgolink/v3/lavalink"
)

var (
	CLIENT = Client{}
	CONFIG = Config{}
	TRACKS = Tracks{
		store: make([]lavalink.Track, 0, 10),
		mu:    sync.Mutex{},
	}

	MENTION = ""

	// For LLM
	HTTP = http.Client{}

	REGISTER = flag.Bool("register", false, "Register commands globally")

	ASSETS_PATHS = map[string]string{}
	ASSETS       = []string{}
)

func main() {
	flag.Parse()

	err := NewConfig()
	if err != nil {
		panic(err)
	}

	err = NewClient()
	if err != nil {
		panic(err)
	}

	go func() {
		for request := range LLM_QUEUE {
			fmt.Printf("Processing LLM request, message ID: %s\n", request.Message.ID)
			processLLM(request)
		}
	}()

	fmt.Println("Client is ready")
	fmt.Println("Press Ctrl+C to exit")

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}
