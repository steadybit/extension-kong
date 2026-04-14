// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import (
	"context"
	"github.com/kong/go-kong/kong"
	"github.com/steadybit/extension-kong/v2/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func getTestService() *kong.Service {
	return &kong.Service{
		Enabled:  new(true),
		Host:     new("mockbin.org"),
		Name:     new("mockbin"),
		Path:     new("/request"),
		Port:     new(443),
		Protocol: new("https"),
	}
}

func getTestRoute(service *kong.Service) *kong.Route {
	return &kong.Route{
		Name:    new("test"),
		Service: service,
		Hosts:   []*string{new("server1")},
		Paths:   []*string{new("/products")},
		Tags:    []*string{new("test")},
		Methods: []*string{new("GET")},
	}
}

func configureService(t *testing.T, instance *config.Instance, service *kong.Service) *kong.Service {
	client, err := instance.GetClient()
	require.NoError(t, err)

	createdService, err := client.Services.Create(context.Background(), service)
	require.NoError(t, err)
	return createdService
}

func configureRoute(t *testing.T, instance *config.Instance, route *kong.Route) *kong.Route {
	client, err := instance.GetClient()
	require.NoError(t, err)

	createdRoute, err := client.Routes.Create(context.Background(), route)
	require.NoError(t, err)
	return createdRoute
}

func getTestConsumer() *kong.Consumer {
	return &kong.Consumer{
		Username: new("test-consumer"),
	}
}

func configureConsumer(t *testing.T, instance *config.Instance, consumer *kong.Consumer) *kong.Consumer {
	client, err := instance.GetClient()
	require.NoError(t, err)

	createdConsumer, err := client.Consumers.Create(context.Background(), consumer)
	require.NoError(t, err)
	return createdConsumer
}
