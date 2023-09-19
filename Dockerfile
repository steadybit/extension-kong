# syntax=docker/dockerfile:1

##
## Build
##
FROM goreleaser/goreleaser:v1.19.2 AS build

ARG BUILD_WITH_COVERAGE
ARG BUILD_SNAPSHOT=true

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN goreleaser build --snapshot="${BUILD_SNAPSHOT}" --single-target -o extension
##
## Runtime
##
FROM alpine:3.17

LABEL "steadybit.com.discovery-disabled"="true"

ARG USERNAME=steadybit
ARG USER_UID=10000

RUN adduser -u $USER_UID -D $USERNAME

USER $USERNAME

WORKDIR /

COPY --from=build /app/extension /extension
COPY --from=build /app/licenses /licenses

EXPOSE 8084 8085

ENTRYPOINT ["/extension"]
