package main

import (
	"cloud.google.com/go/storage"
	"context"
	"github.com/labstack/gommon/log"
	"io"
	"strings"
)

func main(){
	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.

	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	k := strings.NewReader("{}")

	// Upload an object with storage.Writer.
	wc := client.Bucket("floxy-binary").Object("object.json").NewWriter(ctx)
	if _, err = io.Copy(wc, k); err != nil {
		panic(err)
	}
	if err := wc.Close(); err != nil {
		panic(err)
	}

}
