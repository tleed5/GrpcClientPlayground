package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"time"

	pb "GrpcClientPlayground/protos"
	"google.golang.org/grpc"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:8443", "the address to connect to")
	//addr = flag.String("addr", "host.docker.internal:8443", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()

	clientCert, err := tls.LoadX509KeyPair("client-cert.pem", "client-key.pem")
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	config := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{clientCert},
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			return nil
		},
		RootCAs: nil,
	}
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	//conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	doGreet(conn, name)
	//doCountUp(conn)
	//doCountAndReply(conn)
}

// Client to server streaming
func doCountUp(conn *grpc.ClientConn) {
	c := pb.NewCounterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	counter, err := c.AccumulateCount(ctx)
	if err != nil {
		return
	}

	for i := 0; i < 10; i++ {
		err = counter.Send(&pb.CounterRequest{Count: int32(i)})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
	}

	response, err := counter.CloseAndRecv()
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %v", response.GetCount())
}

func doCountAndReply(conn *grpc.ClientConn) {
	c := pb.NewCounterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	stream, err := c.CountAndRespond(ctx)
	if err != nil {
		log.Fatalf("client.RouteChat failed: %v", err)
	}
	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("client.CountAndRespond failed: %v", err)
			}
			log.Printf("Got count %v", in.GetCount())
		}
	}()

	for i := 0; i < 5; i++ {
		log.Printf("Sent count %v", i)
		err = stream.Send(&pb.CounterRequest{Count: int32(i)})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		time.Sleep(time.Second)
	}

	err = stream.CloseSend()
	if err != nil {
		return
	}
	<-waitc
}

func doGreet(conn *grpc.ClientConn, name *string) {
	c := pb.NewGreeterClient(conn)

	for {
		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
		time.Sleep(30 * time.Millisecond)

		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Greeting: %s", r.GetMessage())
		cancel()
	}
}
