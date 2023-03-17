# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.18-alpine AS build

ARG NAME
ARG VERSION
ARG REVISION

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build \
	-ldflags="\
	-X 'github.com/steadybit/extension-kit/extbuild.ExtensionName=${NAME}' \
	-X 'github.com/steadybit/extension-kit/extbuild.Version=${VERSION}' \
	-X 'github.com/steadybit/extension-kit/extbuild.Revision=${REVISION}'" \
	-o /extension-kong

##
## Runtime
##
FROM alpine:3.16

WORKDIR /

COPY --from=build /extension-kong /extension-kong

EXPOSE 8084

ENTRYPOINT ["/extension-kong"]
