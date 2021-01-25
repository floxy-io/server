package main

import (
	"context"
	"github.com/danielsussa/floxy/internal/home"
	"github.com/danielsussa/floxy/internal/infra/db"
	"github.com/danielsussa/floxy/internal/sshserver"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main(){
	ctx := context.Background()

	err := db.Setup()
	if err != nil {
		log.Fatal(err)
	}

	home.Start()
	sshserver.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit

	if err := home.Shutdown(ctx); err != nil {
		log.Fatal(ctx, err.Error())
	}
	if err := sshserver.Shutdown(ctx); err != nil {
		log.Fatal(ctx, err.Error())
	}

}
