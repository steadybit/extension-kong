<img src="./Kong_2x.png" width="300" align="right" alt="Kong logo">

# Steadybit extension-kong

A [steadybit](https://www.steadybit.com/) attack implementation to inject HTTP faults into [Kong API gateway](https://konghq.com/).

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
