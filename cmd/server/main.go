package main

import (
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	pb "ojichat-stream/proto/gen"

	ojc "github.com/greymd/ojichat/generator"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedOjichatServer
}

func receive(ch chan<- pb.ChatRequest, stream pb.Ojichat_ChatServer) {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			continue
		}
		if err != nil {
			log.Println(err)
			return
		}
		ch <- *in
	}
}

func reply(ch chan<- bool, stream pb.Ojichat_ChatServer) {
	for {
		time.Sleep(time.Second * time.Duration(rand.Intn(10)))
		ch <- true
	}
}

func (s *server) Chat(stream pb.Ojichat_ChatServer) error {
	ojiConf := ojc.Config{EmojiNum: rand.Intn(10), PunctuationLevel: rand.Intn(3)}
	recvCh := make(chan pb.ChatRequest)
	go receive(recvCh, stream)

	replyCh := make(chan bool)
	go reply(replyCh, stream)

	for {
		select {
		case v := <-recvCh:
			name := v.GetName()
			msg := v.GetMessage()
			log.Printf("name: %v, message: %v ", name, msg)
			ojiConf.TargetName = name
			if msg == "" {
				continue
			}
			if err := stream.Send(&pb.ChatResponse{Message: "返信ありがとう (^o^)"}); err != nil {
				return err
			}
		case <-replyCh:
			reply, err := ojc.Start(ojiConf)
			if err != nil {
				return err
			}
			if err := stream.Send(&pb.ChatResponse{Message: reply}); err != nil {
				return err
			}
		}
	}
}

func main() {
	port := ":50051"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterOjichatServer(s, &server{})
	if err != nil {
		log.Fatal(err)
	}

	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
