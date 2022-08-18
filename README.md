<img src="./logo.png" height="130" align="right" alt="Kong logo">

# Steadybit extension-kong

A [Steadybit](https://www.steadybit.com/) attack implementation to inject HTTP faults into [Kong API gateway](https://konghq.com/).

## Prerequisites

- Kong needs to have the [request-termination](https://docs.konghq.com/hub/kong-inc/request-termination/#example-use-cases) plugin installed (typically
	installed by default).

## Configuration

| Environment Variable                                 |                                                                                                  |
|------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| `STEADYBIT_EXTENSION_KONG_INSTANCE_<n>_NAME`         | Name of the kong instance                                                                        |
| `STEADYBIT_EXTENSION_KONG_INSTANCE_<n>_ORIGIN`       | Url of the kong admin interface                                                                  |
| `STEADYBIT_EXTENSION_KONG_INSTANCE_<n>_HEADER_KEY`   | Optional header key to send to the Kong admin API. Typically used for authentication purposes.   |
| `STEADYBIT_EXTENSION_KONG_INSTANCE_<n>_HEADER_VALUE` | Optional header value to send to the Kong admin API. Typically used for authentication purposes. |


## Deployment

We recommend that you deploy the extension with our [official Helm chart](https://github.com/steadybit/helm-charts/tree/main/charts/steadybit-extension-kong).

## Agent Configuration

The Steadybit agent needs to be configured to interact with the Kong extension by adding the following environment variables:

```shell
# Make sure to adapt the URLs and indices in the environment variables names as necessary for your setup

STEADYBIT_AGENT_ACTIONS_EXTENSIONS_0_URL=http://steadybit-extension-kong.steadybit-extension.svc.cluster.local:8084
STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_0_URL=http://steadybit-extension-kong.steadybit-extension.svc.cluster.local:8084
```

When leveraging our official Helm charts, you can set the configuration through additional environment variables on the agent:

```
--set agent.env[0].name=STEADYBIT_AGENT_ACTIONS_EXTENSIONS_0_URL \
--set agent.env[0].value="http://steadybit-extension-kong.steadybit-extension.svc.cluster.local:8084" \
--set agent.env[1].name=STEADYBIT_AGENT_DISCOVERIES_EXTENSIONS_0_URL \
--set agent.env[1].value="http://steadybit-extension-kong.steadybit-extension.svc.cluster.local:8084"
```
