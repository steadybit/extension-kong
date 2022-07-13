# extension-kong

A [Steadybit](https://www.steadybit.com/) attack implementation to inject HTTP faults into [Kong API gateway](https://konghq.com/).

## Prerequisites

- Kong needs to have the [request-termination](https://docs.konghq.com/hub/kong-inc/request-termination/#example-use-cases) plugin installed (typically
	installed by default).

## Configuration

| Environment Variable                           |                                 |
|------------------------------------------------|---------------------------------|
| `STEADYBIT_EXTENSION_KONG_INSTANCE_<n>_NAME`   | Name of the kong instance       |
| `STEADYBIT_EXTENSION_KONG_INSTANCE_<n>_ORIGIN` | Url of the kong admin interface |

## Running as docker container

1. Build the image:
	  ```
		docker build -t steadybit/extension-kong .
		```

2. Run the image:
	  ```
		docker run -d -p 8084:8084 --name extension-kong \
			 -e "STEADYBIT_EXTENSION_KONG_INSTANCE_0_NAME=default" \
			 -e "STEADYBIT_EXTENSION_KONG_INSTANCE_0_ORIGIN=http://kong:8001" \
			 steadybit/extension-kong
		```

## Running Kong

https://docs.konghq.com/gateway/latest/install-and-run/docker/
