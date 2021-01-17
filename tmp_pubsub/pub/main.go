package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
)

func main(){
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "floxy-300919")
	if err != nil {
		panic(err)
	}
	t := client.Topic("create-binary")

	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte("Message"),
	})

	id, err := result.Get(ctx)
	fmt.Println(id, err)
}
