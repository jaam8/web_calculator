//go:build integration
// +build integration

package integration_tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/jaam8/web_calculator/common-lib/gen/auth_service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type testAuthServices struct {
	pgContainer          tc.Container
	redisContainer       tc.Container
	authServiceContainer tc.Container
	network              *tc.DockerNetwork

	authClient auth_service.AuthServiceClient

	ctx    context.Context
	cancel context.CancelFunc
}

func setupTestServices(t *testing.T) *testAuthServices {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	testNetwork, err := network.New(ctx)
	require.NoError(t, err)

	NetworkName := testNetwork.Name

	pgContainer, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:        "postgres:17-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Networks:     []string{NetworkName},
			NetworkAliases: map[string][]string{
				NetworkName: {"postgres"},
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5 * time.Second),
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "1234",
				"POSTGRES_DB":       "web_calculator",
			},
		},
		Started: true,
	})
	require.NoError(t, err)

	redisContainer, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:        "redis:latest",
			ExposedPorts: []string{"6379/tcp"},
			Networks:     []string{NetworkName},
			NetworkAliases: map[string][]string{
				NetworkName: {"redis"},
			},
			WaitingFor: wait.ForLog("Ready to accept connections"),
		},
		Started: true,
	})
	require.NoError(t, err)

	authServiceContainer, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		Started: true,
		ContainerRequest: tc.ContainerRequest{
			Image:        "jaam8/web_calculator:auth_service",
			Networks:     []string{NetworkName},
			ExposedPorts: []string{"50051/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("50051/tcp"),
				wait.ForLog("AUTH_SERVICE listening at :50051"),
			).WithDeadline(120 * time.Second),
			// env vars taken from .env file
		},
	})
	require.NoError(t, err)

	authServiceHost, err := authServiceContainer.Host(ctx)
	require.NoError(t, err)
	authServicePort, err := authServiceContainer.MappedPort(ctx, "50051")
	require.NoError(t, err)

	authServiceAddr := fmt.Sprintf("%s:%d", authServiceHost, authServicePort.Int())

	conn, err := grpc.NewClient(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	// logs
	logs, err := authServiceContainer.Logs(ctx)
	require.NoError(t, err)
	buf := new(bytes.Buffer)
	io.Copy(buf, logs)
	t.Logf("Auth service logs:\n%s", buf.String())
	require.NoError(t, err)

	authClient := auth_service.NewAuthServiceClient(conn)

	return &testAuthServices{
		pgContainer:          pgContainer,
		redisContainer:       redisContainer,
		authServiceContainer: authServiceContainer,
		network:              testNetwork,
		authClient:           authClient,
		ctx:                  ctx,
		cancel:               cancel,
	}
}

func (ts *testAuthServices) cleanup(t *testing.T) {
	t.Helper()

	if ts.authServiceContainer != nil {
		require.NoError(t, ts.authServiceContainer.Terminate(ts.ctx))
	}

	if ts.redisContainer != nil {
		require.NoError(t, ts.redisContainer.Terminate(ts.ctx))
	}

	if ts.pgContainer != nil {
		require.NoError(t, ts.pgContainer.Terminate(ts.ctx))
	}

	if ts.network != nil {
		require.NoError(t, ts.network.Remove(ts.ctx))
	}

	ts.cancel()
}

func TestAuthServiceRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем интеграционный тест в коротком режиме")
	}

	ts := setupTestServices(t)
	defer ts.cleanup(t)

	resp, err := ts.authClient.Register(ts.ctx, &auth_service.RegisterRequest{
		Login:    "gopher_register",
		Password: "Password123!",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.GetUserId())
}

func TestAuthServiceLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем интеграционный тест в коротком режиме")
	}

	ts := setupTestServices(t)
	defer ts.cleanup(t)

	_, err := ts.authClient.Register(ts.ctx, &auth_service.RegisterRequest{
		Login:    "gopher_login",
		Password: "Password123!",
	})
	require.NoError(t, err)

	resp, err := ts.authClient.Login(ts.ctx, &auth_service.LoginRequest{
		Login:    "gopher_login",
		Password: "Password123!",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.GetAccessToken())
	assert.NotEmpty(t, resp.GetRefreshToken())
}

func TestAuthServiceRefreshToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем интеграционный тест в коротком режиме")
	}

	ts := setupTestServices(t)
	defer ts.cleanup(t)

	registerResp, err := ts.authClient.Register(ts.ctx, &auth_service.RegisterRequest{
		Login:    "gopher_refresh",
		Password: "Password123!",
	})

	require.NoError(t, err)
	assert.NotNil(t, registerResp)
	assert.NotEmpty(t, registerResp.GetUserId())

	loginResp, err := ts.authClient.Login(ts.ctx, &auth_service.LoginRequest{
		Login:    "gopher_refresh",
		Password: "Password123!",
	})

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, loginResp.GetAccessToken())
	assert.NotEmpty(t, loginResp.GetRefreshToken())

	refreshResp, err := ts.authClient.Refresh(ts.ctx, &auth_service.RefreshRequest{
		RefreshToken: loginResp.GetRefreshToken(),
	})

	assert.NoError(t, err)
	assert.NotNil(t, refreshResp)
	assert.NotEmpty(t, refreshResp.GetAccessToken())
	assert.NotEmpty(t, refreshResp.GetRefreshToken())
}
