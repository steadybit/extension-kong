# Changelog

## v1.4.1

 - Use more specific Kong API gateway API endpoints to avoid security issues related to forbidden API endpoints. Contributed by [@achoimet](https://github.com/achoimet).

## v1.4.0

 - New discovery for Kong routes. Contributed by [@achoimet](https://github.com/achoimet).
 - New request termination attack for Kong routes. Contributed by [@achoimet](https://github.com/achoimet).

## v1.3.0

 - The log level can now be configured through the `STEADYBIT_LOG_LEVEL` environment variable. Contributed by [@achoimet](https://github.com/achoimet).

## v1.2.0

 - Update `go-kong` and use the new APIs so that plugin creation, updates and deletions happen using Kong API paths that are specific to services, i.e., located under `/services`.

## v1.1.1

 - Raise version of the request termination attack to `v1.1.1` to update the configuration within Steadybit.

## v1.1.0

 - Ability to define response bodies, `Content-Type` and triggers. Contributed by [@achoimet](https://github.com/achoimet).

## v1.0.0

 - Initial release
