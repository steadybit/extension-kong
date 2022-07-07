// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"context"
	"fmt"
	"github.com/kong/go-kong/kong"
	"golang.org/x/exp/slices"
	"os"
)

type Instance struct {
	Name    string `json:"name"`
	BaseUrl string `json:"base_url"`
}

var (
	Instances []Instance
)

func init() {
	name := getInstanceName(0)
	for len(name) > 0 {
		index := len(Instances)
		Instances = append(Instances, Instance{
			name,
			getInstanceOrigin(index),
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

func FindInstanceByName(name string) (*Instance, error) {
	instanceIndex := slices.IndexFunc(Instances, func(i Instance) bool { return i.Name == name })
	if instanceIndex < 0 {
		return nil, fmt.Errorf("not found")
	}
	return &Instances[instanceIndex], nil
}

func (i *Instance) getClient() (*kong.Client, error) {
	return kong.NewClient(&i.BaseUrl, nil)
}

func (i *Instance) FindService(nameOrId *string) (*kong.Service, error) {
	client, err := i.getClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Services.Get(ctx, nameOrId)
}

func (i *Instance) FindConsumer(nameOrId *string) (*kong.Consumer, error) {
	client, err := i.getClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Consumers.Get(ctx, nameOrId)
}

func (i *Instance) CreatePlugin(plugin *kong.Plugin) (*kong.Plugin, error) {
	client, err := i.getClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Plugins.Create(ctx, plugin)
}

func (i *Instance) UpdatePlugin(plugin *kong.Plugin) (*kong.Plugin, error) {
	client, err := i.getClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Plugins.Update(ctx, plugin)
}

func (i *Instance) DeletePlugin(nameOrID *string) error {
	client, err := i.getClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	return client.Plugins.Delete(ctx, nameOrID)
}

func (i *Instance) GetServices() ([]*kong.Service, error) {
	client, err := i.getClient()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Services.ListAll(ctx)
}
