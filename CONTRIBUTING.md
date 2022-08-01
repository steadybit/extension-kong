# Contributing

## Build
docker build -t steadybit/extension-kong .

## Running as docker container

```
docker run -d -p 8084:8084 --name extension-kong \
	 -e "STEADYBIT_EXTENSION_KONG_INSTANCE_0_NAME=default" \
	 -e "STEADYBIT_EXTENSION_KONG_INSTANCE_0_ORIGIN=http://kong:8001" \
	 steadybit/extension-kong
```

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

with these settings an admin api is accessible at http://kong-kong-admin.kong.svc.cluster.local:8001 inside the cluster.


