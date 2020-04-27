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
				id:    "1",
				texts: []string{"オス！オラ、孫悟空", "おら、腹へったぞ"},
			},
			"2": {
				id:    "2",
				texts: []string{"おとうさんを…いじめるなーーーーっ!!!!!", "わたしは正義を愛する者、グレートサイヤマンだ!!!"},
			},
			"3": {
				id:    "3",
				texts: []string{"か～め～か～め～波"},
			},
		},
	}
	server := grpc.NewServer()
	proto.RegisterBackendServer2Server(server, b)
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
	id    string
	texts []string
}

type BackendServer struct {
	useMap map[string]userDetail
}

func (b BackendServer) UserTimeline(ctx context.Context, request *proto.UserTimelineRequest) (*proto.UserTimelineResponse, error) {
	id := request.GetId()
	log.Printf("got id is %v", id)

	data, ok := b.useMap[id]
	if !ok {
		return nil, fmt.Errorf("id %v not found", id)
	}

	return &proto.UserTimelineResponse{
		Id:    data.id,
		Texts: data.texts,
	}, nil
}
