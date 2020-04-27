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

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/tommy-sho/kuma-k8s/proto"
)

const (
	port = "50000"
)

func main() {
	b := &BackendServer{
		useMap: map[string]userDetail{
			"1": {
				id:          "1",
				name:        "劉 備",
				description: "孫 悟空（そん ごくう）は鳥山明の漫画『ドラゴンボール』に登場する架空のキャラクター",
			},
			"2": {
				id:          "2",
				name:        "孫 悟飯",
				description: "孫 悟飯（そん　ごはん）は悟空とチチの息子。長男",
			},
			"3": {
				id:          "3",
				name:        "孫 悟天",
				description: "孫 悟天（そん　ごてん）は悟空とチチの息子。次男",
			},
		},
	}
	server := grpc.NewServer()
	proto.RegisterBackendServerServer(server, b)
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

type userDetail struct {
	id          string
	name        string
	description string
}

type BackendServer struct {
	useMap map[string]userDetail
}

func (b BackendServer) UserDetail(ctx context.Context, request *proto.UserDetailRequest) (*proto.UserDetailResponse, error) {
	id := request.GetId()
	log.Printf("got id is %v", id)

	data, ok := b.useMap[id]
	if !ok {
		return nil, fmt.Errorf("id %s not found", id)
	}

	return &proto.UserDetailResponse{
		Id:          data.id,
		Name:        data.name,
		Description: data.description,
	}, nil
}
