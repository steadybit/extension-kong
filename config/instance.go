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

func (i *Instance) FindConsumer(nameOrId *string) (*kong.Consumer, error) {
	client, err := i.GetClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Consumers.Get(ctx, nameOrId)
}

func (i *Instance) CreatePlugin(plugin *kong.Plugin) (*kong.Plugin, error) {
	client, err := i.GetClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Plugins.CreateForService(ctx, plugin.Service.ID, plugin)
}

func (i *Instance) UpdatePlugin(serviceId *string, plugin *kong.Plugin) (*kong.Plugin, error) {
	client, err := i.GetClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Plugins.UpdateForService(ctx, serviceId, plugin)
}

func (i *Instance) DeletePlugin(serviceId *string, nameOrID *string) error {
	client, err := i.GetClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	return client.Plugins.DeleteForService(ctx, serviceId, nameOrID)
}

func (i *Instance) GetServices() ([]*kong.Service, error) {
	client, err := i.GetClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Services.ListAll(ctx)
}
