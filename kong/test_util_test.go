// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package kong

import (
	"context"
	"github.com/kong/go-kong/kong"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-kong/v2/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func getTestService() *kong.Service {
	return &kong.Service{
		Enabled:  discovery_kit_api.Ptr(true),
		Host:     discovery_kit_api.Ptr("mockbin.org"),
		Name:     discovery_kit_api.Ptr("mockbin"),
		Path:     discovery_kit_api.Ptr("/request"),
		Port:     discovery_kit_api.Ptr(443),
		Protocol: discovery_kit_api.Ptr("https"),
	}
}

func getTestRoute(service *kong.Service) *kong.Route {
	return &kong.Route{
		Name:    discovery_kit_api.Ptr("test"),
		Service: service,
		Hosts:   []*string{discovery_kit_api.Ptr("server1")},
		Paths:   []*string{discovery_kit_api.Ptr("/products")},
		Tags:    []*string{discovery_kit_api.Ptr("test")},
		Methods: []*string{extutil.Ptr("GET")},
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
		Username: discovery_kit_api.Ptr("test-consumer"),
	}
}

func configureConsumer(t *testing.T, instance *config.Instance, consumer *kong.Consumer) *kong.Consumer {
	client, err := instance.GetClient()
	require.NoError(t, err)

	createdConsumer, err := client.Consumers.Create(context.Background(), consumer)
	require.NoError(t, err)
	return createdConsumer
}
