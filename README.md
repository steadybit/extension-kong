<img src="./Kong_2x.png" width="300" align="right" alt="Kong logo">

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

## Running as docker container

```
docker run -d -p 8084:8084 --name extension-kong \
	 -e "STEADYBIT_EXTENSION_KONG_INSTANCE_0_NAME=default" \
	 -e "STEADYBIT_EXTENSION_KONG_INSTANCE_0_ORIGIN=http://kong:8001" \
	 steadybit/extension-kong
```

## Running as Kubernetes Deployment
Be aware to change the environment variables `STEADYBIT_EXTENSION_KONG_INSTANCE_0_NAME`and `STEADYBIT_EXTENSION_KONG_INSTANCE_0_ORIGIN` in the Kubernetes manifest to match your Kong instance.

```shell
kubectl apply -f examples/kubernetes.yml
```

## Build
docker build -t steadybit/extension-kong .

## Running Kong

### Docker Guide:
See https://docs.konghq.com/gateway/latest/install-and-run/docker/

### Kubernetes using Helm:

```
helm upgrade --install --create-namespace --namespace kong -f examples/kong-values.yml kong kong/kong
```

Create Example service and route (with port forward on 8001 and 8000)
```
# create example service
curl -i -X POST \
  --url http://localhost:8001/services/ \
  --data 'name=example-service' \
  --data 'url=http://mockbin.org'
echo ""

# "create example route"
curl -i -X POST \
  --url http://localhost:8001/services/example-service/routes \
  --data 'hosts[]=example.com'
echo ""

# "test route"
curl -i \
  --url http://localhost:8000 \
  -H 'Host: example.com'
```

with this settings an admin api is accessible at http://kong-kong-admin.kong.svc.cluster.local:8001 inside the cluster

