ARG USERNAME=codeclarity

# BUILD IMAGE
FROM golang:1.20 AS template-build

COPY .. /usr/srv/template

RUN ls /usr/srv/template

WORKDIR /usr/srv/template

ENV GOOS=linux GOARCH=amd64

RUN [ "go", "build", "-v" ]


# DEV IMAGE
FROM template-build AS template-dev

COPY --from=template-build /usr/srv/template/template /

CMD /template


# PRODUCTION IMAGE
FROM alpine AS template
ARG USERNAME

RUN adduser -D ${USERNAME}
USER ${USERNAME}

WORKDIR /home/${USERNAME}

COPY --from=template-build --chown=codeclarity:codeclarity /usr/srv/template/template .

ENTRYPOINT [ "./template" ]
