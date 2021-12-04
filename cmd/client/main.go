package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"

	pb "ojichat-stream/proto/gen"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func receive(ch chan<- pb.ChatResponse, stream pb.Ojichat_ChatClient) {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			close(ch)
			return
		}
		if err != nil {
			log.Fatal(err)
		}
		ch <- *in
	}
}

func input(ch chan<- pb.ChatRequest, r io.Reader) {
	s := bufio.NewScanner(r)
	fmt.Printf("\n\x1b[36menter message:\x1b[0m")
	for s.Scan() {
		input := pb.ChatRequest{Message: s.Text()}
		ch <- input
	}
}

func exec(name string, addr string, secure bool) {
	var conn *grpc.ClientConn
	var err error
	if secure {
		tlsCredentials := credentials.NewTLS(&tls.Config{})
		conn, err = grpc.Dial(addr, grpc.WithTransportCredentials(tlsCredentials), grpc.WithBlock())
	} else {
		conn, err = grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	}
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewOjichatClient(conn)
	stream, err := client.Chat(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	inputCh := make(chan pb.ChatRequest)
	go input(inputCh, os.Stdin)

	recvCh := make(chan pb.ChatResponse)
	go receive(recvCh, stream)

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("done")
			stream.CloseSend()
			return
		case v := <-recvCh:
			fmt.Printf("\n\x1b[32mおじさん>\x1b[0m %v\n\n", v.Message)
			fmt.Printf("\n\x1b[36menter message:\x1b[0m")
		case v := <-inputCh:
			v.Name = name
			if err := stream.Send(&v); err != nil {
				log.Fatal(err)
			}
		}
	}
}

var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "gRPC client for Ojichat server",
	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatal(err)
		}
		addr, err := cmd.Flags().GetString("addr")
		if err != nil {
			log.Fatal(err)
		}
		secure, err := cmd.Flags().GetBool("secure")
		if err != nil {
			log.Fatal(err)
		}
		exec(name, addr, secure)
	},
}

func init() {
	rootCmd.Flags().StringP("name", "n", "unknown", "enter your name")
	rootCmd.Flags().StringP("addr", "a", "localhost:50051", "enter server address")
	rootCmd.Flags().BoolP("secure", "s", false, "enable secure access")
	if err := rootCmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
