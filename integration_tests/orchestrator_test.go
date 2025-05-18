//go:build integration
// +build integration

package integration_tests

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jaam8/web_calculator/common-lib/gen/orchestrator"
	"github.com/jaam8/web_calculator/common-lib/postgres"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type testOrchestrator struct {
	pgContainer           tc.Container
	orchestratorContainer tc.Container
	network               *tc.DockerNetwork

	orchestratorClient orchestrator.OrchestratorServiceClient

	ctx    context.Context
	cancel context.CancelFunc
}

func CreateUserRecord(t *testing.T, ctx context.Context, pgContainer tc.Container) string {
	pgHost, err := pgContainer.Host(ctx)
	require.NoError(t, err)
	pgPort, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)
	db, err := postgres.New(ctx, postgres.Config{
		Host:     pgHost,
		Port:     uint16(pgPort.Int()),
		Username: "postgres",
		Password: "1234",
		Database: "web_calculator",
		MaxConns: 10,
	})
	require.NoError(t, err)

	var id string
	query := `INSERT INTO users.users (login, password_hash) 
			  VALUES ($1, $2)
			  RETURNING id`
	err = db.QueryRow(ctx, query, "test_user", "12345678").Scan(&id)
	require.NoError(t, err)

	return id
}

func setupTestServices(t *testing.T) *testOrchestrator {
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

	orchestratorContainer, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		Started: true,
		ContainerRequest: tc.ContainerRequest{
			Image:        "jaam8/web_calculator:orchestrator",
			Networks:     []string{NetworkName},
			ExposedPorts: []string{"50052/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("50052/tcp"),
				wait.ForLog("ORCHESTRATOR listening at :50052"),
			).WithDeadline(120 * time.Second),
			// env vars taken from .env file
		},
	})
	require.NoError(t, err)

	orchestratorHost, err := orchestratorContainer.Host(ctx)
	require.NoError(t, err)
	orchestratorPort, err := orchestratorContainer.MappedPort(ctx, "50052")
	require.NoError(t, err)

	orchestratorAddr := fmt.Sprintf("%s:%d", orchestratorHost, orchestratorPort.Int())

	conn, err := grpc.NewClient(orchestratorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	// logs
	logs, err := orchestratorContainer.Logs(ctx)
	require.NoError(t, err)
	buf := new(bytes.Buffer)
	io.Copy(buf, logs)
	t.Logf("Orchestrator logs:\n%s", buf.String())
	require.NoError(t, err)

	orchestratorClient := orchestrator.NewOrchestratorServiceClient(conn)

	return &testOrchestrator{
		pgContainer:           pgContainer,
		orchestratorContainer: orchestratorContainer,
		orchestratorClient:    orchestratorClient,
		network:               testNetwork,
		ctx:                   ctx,
		cancel:                cancel,
	}
}

func (to *testOrchestrator) cleanup(t *testing.T) {
	t.Helper()

	if to.orchestratorContainer != nil {
		require.NoError(t, to.orchestratorContainer.Terminate(to.ctx))
	}

	if to.pgContainer != nil {
		require.NoError(t, to.pgContainer.Terminate(to.ctx))
	}

	if to.network != nil {
		require.NoError(t, to.network.Remove(to.ctx))
	}

	to.cancel()
}

func TestOrchestratorCalculate(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем интеграционный тест в коротком режиме")
	}

	ts := setupTestServices(t)
	defer ts.cleanup(t)

	userId := CreateUserRecord(t, ts.ctx, ts.pgContainer)

	resp, err := ts.orchestratorClient.Calculate(ts.ctx, &orchestrator.CalculateRequest{
		UserId:     userId,
		Expression: "100 - 500 * (2 + 3) / 5",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.GetId())
	assert.NoError(t, uuid.Validate(resp.GetId()))
}

func TestOrchestratorExpressionById(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем интеграционный тест в коротком режиме")
	}

	ts := setupTestServices(t)
	defer ts.cleanup(t)

	userId := CreateUserRecord(t, ts.ctx, ts.pgContainer)

	calcResp, err := ts.orchestratorClient.Calculate(ts.ctx, &orchestrator.CalculateRequest{
		UserId:     userId,
		Expression: "100 - 500",
	})
	assert.NoError(t, err)
	assert.NotNil(t, calcResp)

	exprId := calcResp.GetId()
	assert.NotEmpty(t, exprId)
	assert.NoError(t, uuid.Validate(exprId))

	resp, err := ts.orchestratorClient.ExpressionById(ts.ctx, &orchestrator.ExpressionByIdRequest{
		UserId: userId,
		Id:     exprId,
	})
	assert.NoError(t, err)
	assert.NotNil(t, calcResp)

	expr := resp.GetExpression()
	assert.NotEmpty(t, expr)
	assert.NotEmpty(t, expr.GetId())
	assert.NoError(t, uuid.Validate(expr.GetId()))
	assert.Equal(t, exprId, expr.GetId())
	assert.Equal(t, "pending", expr.GetStatus())
	assert.Equal(t, float64(0), expr.GetResult())

	taskResp, err := ts.orchestratorClient.GetTask(ts.ctx, &emptypb.Empty{})
	require.NoError(t, err)
	assert.NotNil(t, taskResp)

	task := taskResp.GetTask()

	_, err = ts.orchestratorClient.ResultTask(ts.ctx, &orchestrator.ResultTaskRequest{
		ExpressionId: exprId,
		Id:           task.GetId(),
		Result:       task.GetArg1() - task.GetArg2(),
	})
	require.NoError(t, err)

	time.Sleep(5 * time.Second)
	resp, err = ts.orchestratorClient.ExpressionById(ts.ctx, &orchestrator.ExpressionByIdRequest{
		UserId: userId,
		Id:     exprId,
	})
	assert.NoError(t, err)
	assert.NotNil(t, calcResp)

	expr = resp.GetExpression()
	assert.NotEmpty(t, expr)
	assert.NotEmpty(t, expr.GetId())
	assert.NoError(t, uuid.Validate(expr.GetId()))
	assert.Equal(t, exprId, expr.GetId())
	assert.Equal(t, "done", expr.GetStatus())
	assert.NotNil(t, expr.GetResult())
	assert.Equal(t, float64(-400), expr.GetResult())
}

func TestOrchestratorExpressions(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем интеграционный тест в коротком режиме")
	}

	ts := setupTestServices(t)
	defer ts.cleanup(t)

	userId := CreateUserRecord(t, ts.ctx, ts.pgContainer)

	calcResp, err := ts.orchestratorClient.Calculate(ts.ctx, &orchestrator.CalculateRequest{
		UserId:     userId,
		Expression: "100 - 500 * (2 + 3) / 5",
	})
	assert.NoError(t, err)
	assert.NotNil(t, calcResp)

	exprId1 := calcResp.GetId()
	assert.NotEmpty(t, exprId1)
	assert.NoError(t, uuid.Validate(exprId1))

	calcResp, err = ts.orchestratorClient.Calculate(ts.ctx, &orchestrator.CalculateRequest{
		UserId:     userId,
		Expression: "100 - 500 * (2 + 3) / 5 * 70",
	})
	assert.NoError(t, err)
	assert.NotNil(t, calcResp)

	exprId2 := calcResp.GetId()
	assert.NotEmpty(t, exprId2)
	assert.NoError(t, uuid.Validate(exprId2))

	resp, err := ts.orchestratorClient.Expressions(ts.ctx, &orchestrator.ExpressionsRequest{
		UserId: userId,
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.GetExpressions())

	for _, expr := range resp.GetExpressions() {
		assert.NotEmpty(t, expr.GetId())
		assert.NoError(t, uuid.Validate(expr.GetId()))
		assert.Equal(t, "pending", expr.GetStatus())
		assert.Equal(t, float64(0), expr.GetResult())
	}
}
