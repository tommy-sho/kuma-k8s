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
	flag.StringVar(&gateway, "gateway", "localhost:50001", "gateway address")
	flag.StringVar(&id, "id", "", "input ID")
}

func main() {

	flag.Parse()

	ctx := context.Background()

	gConn, err := grpc.DialContext(ctx, gateway, grpc.WithInsecure())
	if err != nil {
		panic(fmt.Errorf("failed to connect with backend server error : %v ", err))
	}
	c := proto.NewGatewayServerClient(gConn)

	r, err := c.GetUser(ctx, &proto.GetUserRequest{Id: id})
	if err != nil {
		fmt.Println("failed to call Gateway error : ", err)
		return
	}
	fmt.Printf("id: %v\nname: %v\ndescription: %v\ntext: %v\n", r.GetId(), r.GetName(), r.GetDescription(), r.GetTimeline())
}
