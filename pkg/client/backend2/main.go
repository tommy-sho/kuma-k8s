package main

import (
	"context"
	"flag"
	"fmt"

	"google.golang.org/grpc"

	"github.com/tommy-sho/kuma-k8s/proto"
)

var (
	gateway string
	id      string
)

func init() {
	flag.StringVar(&gateway, "port", "localhost:50000", "port")
	flag.StringVar(&id, "id", "", "input ID")
}

func main() {

	flag.Parse()

	ctx := context.Background()

	gConn, err := grpc.DialContext(ctx, gateway, grpc.WithInsecure())
	if err != nil {
		panic(fmt.Errorf("failed to connect with backend server error : %v ", err))
	}
	c := proto.NewBackendServer2Client(gConn)

	r, err := c.UserTimeline(ctx, &proto.UserTimelineRequest{Id: id})
	if err != nil {
		fmt.Println("failed to call Gateway error : ", err)
		return
	}
	fmt.Println(r.GetTexts())
}
