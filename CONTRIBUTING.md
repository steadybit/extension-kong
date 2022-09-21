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


## Contributor License Agreement (CLA)

In order to accept your pull request, we need you to submit a CLA. You only need to do this once. If you are submitting a pull request for the first time, just submit a Pull Request and our CLA Bot will give you instructions on how to sign the CLA before merging your Pull Request.

All contributors must sign an [Individual Contributor License Agreement](https://github.com/steadybit/.github/blob/main/.github/cla/individual-cla.md).

If contributing on behalf of your company, your company must sign a [Corporate Contributor License Agreement](https://github.com/steadybit/.github/blob/main/.github/cla/corporate-cla.md). If so, please contact us via office@steadybit.com.

If for any reason, your first contribution is in a PR created by other contributor, please just add a comment to the PR
with the following text to agree our CLA: "I have read the CLA Document and I hereby sign the CLA".

