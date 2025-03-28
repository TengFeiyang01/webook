package main

import (
	artv1 "github.com/TengFeiyang01/webook/webook/api/proto/gen/article/v1"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

func TestGRPCClient(t *testing.T) {
	cc, err := grpc.NewClient("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := artv1.NewArticleServiceClient(cc)
	resp, err := client.List(context.Background(), &artv1.ListRequest{
		Id:     3,
		Offset: 0,
		Limit:  10,
	})
	require.NoError(t, err)
	t.Log(resp.Arts)
}
