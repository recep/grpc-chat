package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	chatpb "github.com/recep/grpc-chat/proto"
	"google.golang.org/grpc"
)

var wg sync.WaitGroup
var grpcClient chatpb.ChatServiceClient

func main() {
	username := flag.String("n", "A", "username")
	flag.Parse()

	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}

	grpcClient = chatpb.NewChatServiceClient(conn)

	stream, err := grpcClient.CreateConnection(context.Background(), &chatpb.Connection{Username: *username})
	if err != nil {
		log.Fatalln(err)
	}

	wg.Add(1)
	go func(str chatpb.ChatService_CreateConnectionClient) {
		defer wg.Done()
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				continue
			}
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(msg)
		}
	}(stream)

	wg.Add(1)
	go sendMessage(*username)

	wg.Wait()
}

func sendMessage(username string) {
	defer wg.Done()
	ts := time.Now()

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		m := &chatpb.Message{
			Username:  username,
			Content:   scanner.Text(),
			Timestamp: ts.String(),
		}

		_, err := grpcClient.SendMessage(context.Background(), m)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
