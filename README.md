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

## Installation

### Kubernetes

Detailed information about agent and extension installation in kubernetes can also be found in
our [documentation](https://docs.steadybit.com/install-and-configure/install-agent/install-on-kubernetes).

#### Recommended (via agent helm chart)

All extensions provide a helm chart that is also integrated in the
[helm-chart](https://github.com/steadybit/helm-charts/tree/main/charts/steadybit-agent) of the agent.

You must provide additional values to activate this extension.

```
--set extension-kong.enabled=true \
--set extension-kong.kong.name="{{SYMBOLIC_NAME}}" \
--set extension-kong.kong.origin="{{KONG_API_SERVER_ORIGIN}}" \
```

Additional configuration options can be found in
the [helm-chart](https://github.com/steadybit/extension-kong/blob/main/charts/steadybit-extension-kong/values.yaml) of the
extension.

#### Alternative (via own helm chart)

If you need more control, you can install the extension via its
dedicated [helm-chart](https://github.com/steadybit/extension-kong/blob/main/charts/steadybit-extension-kong).

```bash
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

### Linux Package

Please use
our [agent-linux.sh script](https://docs.steadybit.com/install-and-configure/install-agent/install-on-linux-hosts)
to install the extension on your Linux machine. The script will download the latest version of the extension and install
it using the package manager.

After installing, configure the extension by editing `/etc/steadybit/extension-kong` and then restart the service.

## Extension registration

Make sure that the extension is registered with the agent. In most cases this is done automatically. Please refer to
the [documentation](https://docs.steadybit.com/install-and-configure/install-agent/extension-discovery) for more
information about extension registration and how to verify.
