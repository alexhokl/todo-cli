FROM golang:1.26.5-alpine AS builder

RUN apk update && apk upgrade && \
    apk add --no-cache git build-base

# The SQLite driver (mattn/go-sqlite3) is a cgo package, so cgo cannot be
# disabled here. build-base above provides the toolchain it needs.
ENV CGO_ENABLED=1
ENV app_name=todo-cli
ENV repo=github.com/alexhokl/${app_name}

# Supplied by Cloud Build from $TAG_NAME and $SHORT_SHA so that `todo version`
# reports something meaningful in a released image.
ARG GIT_TAG
ARG GIT_COMMIT

WORKDIR /go/src/${repo}

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# The :- defaults are deliberate. Cloud Build substitutes an empty string for
# $TAG_NAME on a non tag triggered build, which would otherwise stamp an empty
# version rather than falling back.
RUN go install -ldflags "\
      -X ${repo}/cmd.version=${GIT_TAG:-dev} \
      -X ${repo}/cmd.commit=${GIT_COMMIT:-none}"

# Alpine 3.24 matches the builder image. The binary links dynamically against
# musl because cgo is required, so the two releases have to agree.
FROM alpine:3.24 AS dev

ENV USERNAME=appuser
ENV UID=1001
ENV GROUP=appgroup
# HOME must stay set. The CLI resolves its configuration directory through
# os.UserConfigDir(), which fails when neither HOME nor XDG_CONFIG_HOME is set,
# and the helper library exits the process when that lookup fails.
ENV HOME=/home/${USERNAME}

# The database lives on a volume so that it survives container replacement.
# Without this the default path resolves under HOME, inside the container
# filesystem, and is lost on restart.
ENV TODO_DATABASE=/mnt/data/todo.db

RUN addgroup -g ${UID} -S ${GROUP} && adduser -u ${UID} -S -G ${GROUP} ${USERNAME}

WORKDIR $HOME

RUN mkdir -p /mnt/data && \
    chown -R ${USERNAME}:${GROUP} /mnt/data $HOME

COPY --from=builder --chown=appuser:appgroup /go/bin/todo-cli .

VOLUME /mnt/data

EXPOSE 8080

ENTRYPOINT ["./todo-cli"]
CMD ["serve"]

FROM dev AS production

USER ${USERNAME}
