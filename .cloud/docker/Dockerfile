# syntax=docker/dockerfile:1

ARG USERNAME=codeclarity

# BUILD IMAGE
FROM golang:1.22.2-alpine AS plugin-build
COPY . /codeclarity
WORKDIR /codeclarity
ADD .cloud/docker/config/.netrc /root/.netrc
RUN apk add git
RUN rm go.work
RUN rm go.work.sum
RUN [ "go", "build", "-o", "plugin" ]

# PRODUCTION IMAGE
FROM alpine:3.19.1 AS plugin
ARG USERNAME
RUN adduser -D ${USERNAME}
USER ${USERNAME}
WORKDIR /home/${USERNAME}/codeclarity
COPY --from=plugin-build --chown=codeclarity:codeclarity /codeclarity/plugin .
COPY --from=plugin-build --chown=codeclarity:codeclarity /codeclarity/config.json .

ENTRYPOINT [ "./plugin" ]