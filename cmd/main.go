package main

import (
	"context"
	"fmt"
	"github.com/danielsussa/floxy/internal/entrypoints/httpserver"
	"github.com/danielsussa/floxy/internal/entrypoints/sshserver"
	"github.com/danielsussa/floxy/internal/pkg/store"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// using example: sshpass -p testing ssh -N -R 0:localhost:1323 localhost -p 2222
func main() {
	ctx := context.Background()
	log.Println("start log!")

	e := store.New()

	sshserver.Start(e)
	<-httpserver.Start(e)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-quit:
		fmt.Println("quit")
		break
		//case <- httpErr:
		//	fmt.Println("http err")
		//	break
	}

	if err := sshserver.Shutdown(ctx); err != nil {
		log.Fatal(ctx, err.Error())
	}
	if err := httpserver.Shutdown(); err != nil {
		log.Fatal(ctx, err.Error())
	}

}
