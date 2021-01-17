package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
)

func main()  {
	// projectID := "my-project-id"
	// subID := "my-sub"
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "floxy-300919")
	if err != nil {
		panic(err)
	}
	sub := client.Subscription("create-binary-sub")
	cctx, _ := context.WithCancel(ctx)
	err = sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		fmt.Println("Got message: ", string(msg.Data))
		msg.Ack()
	})
	if err != nil {
		panic(err)
	}
}
