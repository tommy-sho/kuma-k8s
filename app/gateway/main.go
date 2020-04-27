package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/tommy-sho/kuma-k8s/proto"
)

const (
	port = "50000"
)

var (
	backend1Host = "localhost:50001"
	backend2Host = "localhost:50002"
)

func init() {
	if host := os.Getenv("BACKEND"); host != "" {
		backend1Host = host
	}

	if host := os.Getenv("BACKEND2"); host != "" {
		backend2Host = host
	}
}

func main() {
	ctx := context.Background()
	callOption := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	bConn, err := grpc.DialContext(ctx, backend1Host, []grpc.DialOption{
		grpc.WithInsecure(), grpc.WithDefaultCallOptions(callOption...),
	}...)
	if err != nil {
		panic(fmt.Errorf("failed to connect with backend server error : %v ", err))
	}
	log.Printf("backend host is %v\n", backend1Host)
	bClient := proto.NewBackendServerClient(bConn)

	b2Conn, err := grpc.DialContext(ctx, backend2Host, []grpc.DialOption{
		grpc.WithInsecure(), grpc.WithDefaultCallOptions(callOption...),
	}...)
	if err != nil {
		panic(fmt.Errorf("failed to connect with backend server error : %v ", err))
	}

	b2Client := proto.NewBackendServer2Client(b2Conn)

	b := &GatewayServer{
		backendClient:  bClient,
		backend2Client: b2Client,
	}

	server := grpc.NewServer()
	proto.RegisterGatewayServerServer(server, b)
	reflection.Register(server)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		panic(err)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	go func() {
		<-stopChan
		gracefulStopChan := make(chan bool, 1)
		go func() {
			server.GracefulStop()
			gracefulStopChan <- true
		}()
		t := time.NewTimer(10 * time.Second)
		select {
		case <-gracefulStopChan:
			log.Print("Success graceful stop")
		case <-t.C:
			server.Stop()
		}
	}()

	errors := make(chan error)
	go func() {
		log.Print("start server. port is ", port)
		errors <- server.Serve(lis)
	}()

	if err := <-errors; err != nil {
		log.Fatal("Failed to server gRPC server", err)
	}

}

type GatewayServer struct {
	backendClient  proto.BackendServerClient
	backend2Client proto.BackendServer2Client
}

func (g *GatewayServer) GetUser(ctx context.Context, request *proto.GetUserRequest) (*proto.GetUserResponse, error) {
	var (
		id         = request.GetId()
		eg, newCtx = errgroup.WithContext(ctx)

		res1 *proto.UserDetailResponse
		res2 *proto.UserTimelineResponse
	)

	log.Printf("got id is %v", id)
	eg.Go(func() error {
		res, err := g.backendClient.UserDetail(newCtx, &proto.UserDetailRequest{Id: id})
		if err != nil {
			fmt.Println("error1：", err)
			return err
		}
		fmt.Println("res1: ", res)

		res1 = res
		return nil
	})

	eg.Go(func() error {
		res, err := g.backend2Client.UserTimeline(newCtx, &proto.UserTimelineRequest{Id: id})
		if err != nil {
			fmt.Println("error2：", err)
			return err
		}
		fmt.Println("res2: ", res)
		res2 = res
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf(":%w", err)
	}

	return &proto.GetUserResponse{
		Id:          id,
		Name:        res1.GetName(),
		Description: res1.GetDescription(),
		Timeline:    res2.GetTexts(),
	}, nil
}
