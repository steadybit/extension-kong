// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/extension-kong/v2/config"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

type TestContainers struct {
	PostgresContainer *testcontainers.Container
	KongContainer     *testcontainers.Container
	Instance          *config.Instance
}

func (tcs *TestContainers) Terminate(t *testing.T, ctx context.Context) {
	log.Info().Msgf("Terminating Kong gateway container")
	kongContainer := *tcs.KongContainer
	err := kongContainer.Terminate(ctx)
	require.NoError(t, err)

	log.Info().Msgf("Terminating Postgres container")
	postgresContainer := *tcs.PostgresContainer
	err = postgresContainer.Terminate(ctx)
	require.NoError(t, err)
}

type WithTestContainersCase struct {
	Name string
	Test func(t *testing.T, instance *config.Instance)
}

func WithTestContainers(t *testing.T, testCases []WithTestContainersCase) {
	tcs, err := setupTestContainers(context.Background())
	require.NoError(t, err)
	defer tcs.Terminate(t, context.Background())

	instance := tcs.Instance
	config.Instances = append(config.Instances, *instance)
	defer resetGlobalInstanceConfiguration()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			defer cleanupKong(t, instance)
			tc.Test(t, instance)
		})
	}
}

func resetGlobalInstanceConfiguration() {
	config.Instances = []config.Instance{}
}

func cleanupKong(t *testing.T, instance *config.Instance) {
	client, err := instance.GetClient()
	require.NoError(t, err)

	// delete all routes
	routes, err := client.Routes.ListAll(context.Background())
	require.NoError(t, err)
	for _, route := range routes {
		err = client.Routes.Delete(context.Background(), route.ID)
		require.NoError(t, err)
	}

	// delete all services
	services, err := client.Services.ListAll(context.Background())
	require.NoError(t, err)
	for _, service := range services {
		err = client.Services.Delete(context.Background(), service.ID)
		require.NoError(t, err)
	}

	// delete all plugins
	plugins, err := client.Plugins.ListAll(context.Background())
	require.NoError(t, err)
	for _, plugin := range plugins {
		err = client.Plugins.Delete(context.Background(), plugin.ID)
		require.NoError(t, err)
	}

	// delete all consumers
	consumers, err := client.Consumers.ListAll(context.Background())
	require.NoError(t, err)
	for _, consumer := range consumers {
		err = client.Consumers.Delete(context.Background(), consumer.ID)
		require.NoError(t, err)
	}
}

func setupTestContainers(ctx context.Context) (*TestContainers, error) {
	kongImage := "kong/kong-gateway:2.8.1.2-alpine"

	network, err := tcnetwork.New(ctx)
	if err != nil {
		return nil, err
	}

	postgresReq := testcontainers.ContainerRequest{
		Image: "postgres:9.6",
		Name:  "test-kong-database",
		Env: map[string]string{
			"POSTGRES_USER":     "kong",
			"POSTGRES_DB":       "kong",
			"POSTGRES_PASSWORD": "kongpass",
		},
		ExposedPorts: []string{"5433/tcp"},
		Networks:     []string{network.Name},
		Cmd:          []string{"-p", "5433"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 10)

	kongMigratorReq := testcontainers.ContainerRequest{
		Image: kongImage,
		Env: map[string]string{
			"KONG_DATABASE":    "postgres",
			"KONG_PG_HOST":     "test-kong-database",
			"KONG_PG_PORT":     "5433",
			"KONG_PG_USER":     "kong",
			"KONG_PG_PASSWORD": "kongpass",
		},
		Networks:   []string{network.Name},
		Cmd:        []string{"kong", "migrations", "bootstrap", "-v"},
		WaitingFor: wait.ForExit(),
	}
	_, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: kongMigratorReq,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	kongReq := testcontainers.ContainerRequest{
		Image: kongImage,
		Name:  "test-kong-gateway",
		ExposedPorts: []string{
			// gateway ingress port
			"8000/tcp",
			// gateway admin API port
			"8001/tcp",
			// gateway UI port
			"8002/tcp"},
		Env: map[string]string{
			"KONG_DATABASE":         "postgres",
			"KONG_PG_HOST":          "test-kong-database",
			"KONG_PG_PORT":          "5433",
			"KONG_PG_USER":          "kong",
			"KONG_PG_PASSWORD":      "kongpass",
			"KONG_PROXY_ACCESS_LOG": "/dev/stdout",
			"KONG_ADMIN_ACCESS_LOG": "/dev/stdout",
			"KONG_PROXY_ERROR_LOG":  "/dev/stderr",
			"KONG_ADMIN_ERROR_LOG":  "/dev/stderr",
			"KONG_ADMIN_LISTEN":     "0.0.0.0:8001",
			"KONG_ADMIN_GUI_URL":    "http://localhost:8002",
		},
		Networks:   []string{network.Name},
		WaitingFor: wait.ForHTTP("/").WithPort("8001/tcp").WithStartupTimeout(5 * time.Minute),
	}
	kongContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: kongReq,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	kongIp, err := kongContainer.Host(ctx)
	if err != nil {
		return nil, err
	}

	kongPort, err := kongContainer.MappedPort(ctx, "8001")
	if err != nil {
		return nil, err
	}

	kongOrigin := fmt.Sprintf("http://%s:%s", kongIp, kongPort.Port())

	return &TestContainers{
		PostgresContainer: &postgresContainer,
		KongContainer:     &kongContainer,
		Instance:          &config.Instance{Name: "test-local", BaseUrl: kongOrigin},
	}, nil
}
