# Changelog

## v2.0.4

- update dependencies

## v2.0.3

- migration to new unified steadybit actionIds and targetTypes

## v2.0.2

- update dependencies

## v2.0.1

- update dependencies

## v2.0.1

 - add linux package build

## v2.0.0

 - Refactoring using `action_kit_sdk`
 - Read only file system

## v1.7.0

 - Print build information on extension startup.

## v1.6.1

 - Fix an issue that can occur with routes without an ID or name. Contributed by [@achoimet](https://github.com/achoimet).

## v1.6.0

 - Support creation of a TLS server through the environment variables `STEADYBIT_EXTENSION_TLS_SERVER_CERT` and `STEADYBIT_EXTENSION_TLS_SERVER_KEY`. Both environment variables must refer to files containing the certificate and key in PEM format.
 - Support mutual TLS through the environment variable `STEADYBIT_EXTENSION_TLS_CLIENT_CAS`. The environment must refer to a comma-separated list of files containing allowed clients' CA certificates in PEM format.

## v1.5.0

- Support for the `STEADYBIT_LOG_FORMAT` env variable. When set to `json`, extensions will log JSON lines to stderr.

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
