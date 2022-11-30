# Copyright © 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM golang:1.18-alpine3.16 AS builder

ENV GO111MODULE=on
WORKDIR /device-scale

LABEL license='SPDX-License-Identifier: Apache-2.0' \
    copyright='Copyright (c) 2018 - 2022: Intel'

# add git for go modules
RUN apk update && apk add --no-cache make git

COPY go.mod .

RUN go mod download

COPY . .

RUN go mod tidy
RUN go build -tags no_zmq -o device-scale cmd/main.go

# Next image - Copy built Go binary into new workspace
FROM alpine:3.16

LABEL license='SPDX-License-Identifier: Apache-2.0' \
    copyright='Copyright (c) 2018 - 2022: Intel'

WORKDIR /
COPY --from=builder /device-scale/device-scale /device-scale
COPY --from=builder /device-scale/cmd/res/configuration.toml /res/configuration.toml
COPY --from=builder /device-scale/cmd/res/profiles/cas-scale.yaml /res/profiles/cas-scale.yaml
COPY --from=builder /device-scale/cmd/res/devices/device-list.toml /res/devices/device-list.toml

CMD [ "/device-scale","--cp=consul://edgex-core-consul:8500","--registry","--confdir=/res"]
