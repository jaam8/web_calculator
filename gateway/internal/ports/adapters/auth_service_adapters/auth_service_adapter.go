package auth_service_adapters

import (
	"context"
	"fmt"
	auth "github.com/jaam8/web_calculator/common-lib/gen/auth_service"
	"github.com/jaam8/web_calculator/common-lib/grpc/pool"
)

type AuthServiceAdapter struct {
	grpcPool *pool.GrpcPool
}

func NewAuthServiceAdapter(grpcPool *pool.GrpcPool) *AuthServiceAdapter {
	return &AuthServiceAdapter{
		grpcPool: grpcPool,
	}
}

func (a AuthServiceAdapter) Login(request *auth.LoginRequest) (*auth.LoginResponse, error) {
	conn, err := a.grpcPool.GetConn()
	if err != nil {
		return nil, fmt.Errorf("couldn't get conn from pool: %w", err)
	}
	defer conn.Close()             //nolint
	defer a.grpcPool.Restore(conn) //nolint
	client := auth.NewAuthServiceClient(conn)
	response, grpcErr := client.Login(context.Background(), request)
	if grpcErr != nil {
		return nil, fmt.Errorf("error in Login grpc: %w", grpcErr)
	}
	return response, nil
}

func (a AuthServiceAdapter) Register(request *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	conn, err := a.grpcPool.GetConn()
	if err != nil {
		return nil, fmt.Errorf("couldn't get conn from pool: %w", err)
	}
	defer conn.Close()             //nolint
	defer a.grpcPool.Restore(conn) //nolint
	client := auth.NewAuthServiceClient(conn)
	response, grpcErr := client.Register(context.Background(), request)
	if grpcErr != nil {
		return nil, fmt.Errorf("error in Register grpc: %w", grpcErr)
	}
	return response, nil
}

func (a AuthServiceAdapter) Refresh(request *auth.RefreshRequest) (*auth.RefreshResponse, error) {
	conn, err := a.grpcPool.GetConn()
	if err != nil {
		return nil, fmt.Errorf("couldn't get conn from pool: %w", err)
	}
	defer conn.Close()             //nolint
	defer a.grpcPool.Restore(conn) //nolint
	client := auth.NewAuthServiceClient(conn)
	response, grpcErr := client.Refresh(context.Background(), request)
	if grpcErr != nil {
		return nil, fmt.Errorf("error in Refresh grpc: %w", grpcErr)
	}
	return response, nil
}
