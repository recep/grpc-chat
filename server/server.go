package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	chatpb "github.com/recep/grpc-chat/proto"
	"google.golang.org/grpc"
)

type server struct {
	connections []*connection
}

type connection struct {
	username string
	stream   chatpb.ChatService_CreateConnectionServer
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	var conns []*connection
	srv := &server{conns}

	s := grpc.NewServer()
	chatpb.RegisterChatServiceServer(s, srv)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

func (s *server) CreateConnection(conn *chatpb.Connection, stream chatpb.ChatService_CreateConnectionServer) error {
	nConn := &connection{
		username: conn.Username,
		stream:   stream,
	}
	s.connections = append(s.connections, nConn)

	//
	for {
		err := stream.Send(&chatpb.Message{
			Content:   "",
			Username:  "computer",
			Timestamp: "",
		})
		if err != nil {
			return err
		}
		time.Sleep(time.Second * 3)
	}
	//

	return nil
}

func (s *server) SendMessage(ctx context.Context, msg *chatpb.Message) (*chatpb.Status, error) {
	wg := &sync.WaitGroup{}

	for _, conn := range s.connections {
		wg.Add(1)
		go func(msg *chatpb.Message, conn *connection) {
			defer wg.Done()

			err := conn.stream.Send(msg)
			if err != nil {
				log.Fatalln(err)
			}
		}(msg, conn)
	}

	wg.Wait()
	return &chatpb.Status{}, nil
}
