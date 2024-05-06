<img src="./logo.png" height="130" align="right" alt="Kong logo">

# Steadybit extension-kong

A [Steadybit](https://www.steadybit.com/) attack implementation to inject HTTP faults into [Kong API gateway](https://konghq.com/).

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.steadybit.extension_kong).

## Prerequisites

- Kong needs to have the [request-termination](https://docs.konghq.com/hub/kong-inc/request-termination/#example-use-cases) plugin installed (typically
	installed by default).

## Configuration

| Environment Variable                                        | Helm value                              | Meaning                                                                                                                | required |
|-------------------------------------------------------------|-----------------------------------------|------------------------------------------------------------------------------------------------------------------------|----------|
| `STEADYBIT_EXTENSION_KONG_INSTANCE_<n>_NAME`                | `kong.name`                             | Name of the kong instance                                                                                              | yes      |
| `STEADYBIT_EXTENSION_KONG_INSTANCE_<n>_ORIGIN`              | `kong.origin`                           | Url of the kong admin interface                                                                                        | yes      |
| `STEADYBIT_EXTENSION_KONG_INSTANCE_<n>_HEADER_KEY`          | `kong.headerKey`                        | Optional header key to send to the Kong admin API. Typically used for authentication purposes.                         | no       |
| `STEADYBIT_EXTENSION_KONG_INSTANCE_<n>_HEADER_VALUE`        | `kong.headerValue`                      | Optional header value to send to the Kong admin API. Typically used for authentication purposes.                       | no       |
| `STEADYBIT_EXTENSION_DISCOVERY_ATTRIBUTES_EXCLUDES_SERVICE` | `discovery.attributes.excludes.service` | List of Target Attributes which will be excluded during discovery. Checked by key equality and supporting trailing "*" | no       |
| `STEADYBIT_EXTENSION_DISCOVERY_ATTRIBUTES_EXCLUDES_ROUTE`   | `discovery.attributes.excludes.route`   | List of Target Attributes which will be excluded during discovery. Checked by key equality and supporting trailing "*" | no       |

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

When installed as linux package this configuration is in`/etc/steadybit/extension-kong`.

## Installation

We recommend that you deploy the extension with our [official Helm chart](https://github.com/steadybit/extension-kong/tree/main/charts/steadybit-extension-kong).

### Helm

```sh
helm repo add steadybit-extension-kong https://steadybit.github.io/extension-kong
helm repo update

helm upgrade steadybit-extension-kong \
  --install \
  --wait \
  --timeout 5m0s \
  --create-namespace \
  --namespace steadybit-agent \
  --set kong.name="{{SYMBOLIC_NAME}}" \
  --set kong.origin="{{KONG_API_SERVER_ORIGIN}}" \
  steadybit-extension-kong/steadybit-extension-kong
```

### Docker

You may alternatively start the Docker container manually.

```sh
docker run \
  --env STEADYBIT_LOG_LEVEL=info \
  --env STEADYBIT_LOG_LEVEL=info \
  --env STEADYBIT_EXTENSION_KONG_INSTANCE_0_ORIGIN="{{KONG_API_SERVER_ORIGIN}}" \
  --expose 8084 \
  ghcr.io/steadybit/extension-kong:latest
```

### Linux Package

Please use our [agent-linux.sh script](https://docs.steadybit.com/install-and-configure/install-agent/install-on-linux-hosts) to install the extension on your Linux machine.
The script will download the latest version of the extension and install it using the package manager.

After installing configure the extension by editing `/etc/steadybit/extension-kong` and then restart the service.

## Register the extension

Make sure to register the extension at the steadybit platform. Please refer to
the [documentation](https://docs.steadybit.com/integrate-with-steadybit/extensions/extension-installation) for more information.
