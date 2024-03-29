// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package config

import (
	"context"
	"fmt"
	"github.com/kong/go-kong/kong"
	"os"
)

type Instance struct {
	Name        string `json:"name"`
	BaseUrl     string `json:"baseUrl"`
	HeaderKey   string `json:"headerKey"`
	HeaderValue string `json:"headerValue"`
}

var (
	Instances []Instance
)

func init() {
	name := getInstanceName(0)
	for len(name) > 0 {
		index := len(Instances)
		Instances = append(Instances, Instance{
			Name:        name,
			BaseUrl:     getInstanceOrigin(index),
			HeaderKey:   getAuthHeaderKey(index),
			HeaderValue: getAuthHeaderValue(index),
		})
		name = getInstanceName(len(Instances))
	}
}

func getInstanceName(n int) string {
	return os.Getenv(fmt.Sprintf("STEADYBIT_EXTENSION_KONG_INSTANCE_%d_NAME", n))
}

func getInstanceOrigin(n int) string {
	return os.Getenv(fmt.Sprintf("STEADYBIT_EXTENSION_KONG_INSTANCE_%d_ORIGIN", n))
}

func getAuthHeaderKey(n int) string {
	return os.Getenv(fmt.Sprintf("STEADYBIT_EXTENSION_KONG_INSTANCE_%d_HEADER_KEY", n))
}

func getAuthHeaderValue(n int) string {
	return os.Getenv(fmt.Sprintf("STEADYBIT_EXTENSION_KONG_INSTANCE_%d_HEADER_VALUE", n))
}

func FindInstanceByName(name string) (*Instance, error) {
	for _, i := range Instances {
		if i.Name == name {
			return &i, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (i *Instance) IsAuthenticated() bool {
	return len(i.HeaderKey) > 0 && len(i.HeaderValue) > 0
}

func (i *Instance) GetClient() (*kong.Client, error) {
	headers := map[string][]string{
		"User-Agent": {"steadybit-extension-kong"},
	}

	if i.IsAuthenticated() {
		headers[i.HeaderKey] = []string{i.HeaderValue}
	}

	client := kong.HTTPClientWithHeaders(nil, headers)
	return kong.NewClient(&i.BaseUrl, client)
}

func (i *Instance) FindService(nameOrId *string) (*kong.Service, error) {
	client, err := i.GetClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Services.Get(ctx, nameOrId)
}

func (i *Instance) FindRoute(service *kong.Service, nameOrId *string) (*kong.Route, error) {
	client, err := i.GetClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	routes, _, err := client.Routes.ListForService(ctx, service.ID, nil)
	if err != nil {
		return nil, err
	}
	var routeFound *kong.Route
	for _, route := range routes {
		if *route.ID == *nameOrId || *route.Name == *nameOrId {
			routeFound = route
		}
	}
	if routeFound == nil {
		return nil, fmt.Errorf("the route %s does not belong to the service %s", *nameOrId, *service.Name)
	}
	return routeFound, nil
}

func (i *Instance) FindConsumer(nameOrId *string) (*kong.Consumer, error) {
	client, err := i.GetClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Consumers.Get(ctx, nameOrId)
}

func (i *Instance) CreatePluginAtAnyLevel(plugin *kong.Plugin) (*kong.Plugin, error) {
	client, err := i.GetClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	if plugin.Route != nil {
		return client.Plugins.CreateForRoute(ctx, plugin.Route.ID, plugin)
	}

	return client.Plugins.CreateForService(ctx, plugin.Service.ID, plugin)
}

func (i *Instance) UpdatePluginForService(serviceId *string, plugin *kong.Plugin) (*kong.Plugin, error) {
	client, err := i.GetClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Plugins.UpdateForService(ctx, serviceId, plugin)
}

func (i *Instance) UpdatePluginForRoute(routeId *string, plugin *kong.Plugin) (*kong.Plugin, error) {
	client, err := i.GetClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Plugins.UpdateForRoute(ctx, routeId, plugin)
}

func (i *Instance) DeletePluginForService(serviceId *string, nameOrID *string) error {
	client, err := i.GetClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	return client.Plugins.DeleteForService(ctx, serviceId, nameOrID)
}

func (i *Instance) DeletePluginForRoute(routeId *string, nameOrID *string) error {
	client, err := i.GetClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	return client.Plugins.DeleteForRoute(ctx, routeId, nameOrID)
}

func (i *Instance) GetServices() ([]*kong.Service, error) {
	client, err := i.GetClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Services.ListAll(ctx)
}

func (i *Instance) GetRoutesForService(serviceNameOrID *string) ([]*kong.Route, *kong.ListOpt, error) {
	client, err := i.GetClient()
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	return client.Routes.ListForService(ctx, serviceNameOrID, nil)
}
