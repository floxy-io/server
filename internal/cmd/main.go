package main

import (
	"context"
	"fmt"
	"github.com/danielsussa/floxy/internal/home"
	"github.com/danielsussa/floxy/internal/httpserver"
	"github.com/danielsussa/floxy/internal/infra/db"
	"github.com/danielsussa/floxy/internal/sshserver"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main(){
	ctx := context.Background()
	//startLog()
	log.Println("start log!")

	err := db.Setup()
	if err != nil {
		log.Fatal(err)
	}

	home.Start()
	sshserver.Start()
	httpErr := httpserver.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-quit:
		fmt.Println("quit")
		break
	case <- httpErr:
		fmt.Println("http err")
		break
	}


	if err := home.Shutdown(ctx); err != nil {
		log.Fatal(ctx, err.Error())
	}
	if err := sshserver.Shutdown(ctx); err != nil {
		log.Fatal(ctx, err.Error())
	}
	if err := httpserver.Shutdown(); err != nil {
		log.Fatal(ctx, err.Error())
	}

}

func startLog(){
	file, err := os.OpenFile("build/floxy.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	log.SetOutput(file)
}