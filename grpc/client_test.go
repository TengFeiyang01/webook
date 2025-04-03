package grpc

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	cc, err := grpc.Dial("localhost:8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.GetById(ctx, &GetByIdReq{
		Id: 456,
	})
	assert.NoError(t, err)
	t.Log(resp.User)
}
