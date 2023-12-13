#
# 1. Build Container
#
FROM golang:1.20 AS build
ARG VERSION
ARG CI_JOB_TOKEN
ARG CI_SERVER_HOST
ARG GITHUB_PAT

COPY archway /home/src/archway
COPY go.mod go.sum Makefile /home/src/
WORKDIR /home/src
RUN go mod download
COPY . /home/src/

# Build components.
RUN make build

#
# 2. Runtime Container
#
FROM alpine as archway-publisher

ENV TZ=Europe/Vilnius \
    PATH="/app:${PATH}"

RUN apk add --update --no-cache \
    tzdata \
    ca-certificates \
    bash \
    && \
    cp --remove-destination /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo "${TZ}" > /etc/timezone

WORKDIR /home/app
COPY --from=build /home/src/dist/archway-publisher .
COPY --from=build /home/src/dist/libwasmvm.x86_64.so .

CMD ["sh", "-c", "LDLIBRARY_PATH=. ./archway-publisher --nats ${NATS_HOSTS} --nats-jwt ${NATS_JWT} --nats-nkey ${NATS_NKEY} start"]