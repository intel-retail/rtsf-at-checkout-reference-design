# Copyright © 2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM golang:1.20-alpine3.17 AS builder

ENV GO111MODULE=on
WORKDIR /device-scale

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
    copyright='Copyright (c) 2018 - 2023: Intel'

# add git for go modules
RUN apk update && apk add --no-cache make git

COPY go.mod .

RUN go mod download

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 go build -tags no_zmq -o device-scale cmd/main.go

# Next image - Copy built Go binary into new workspace
FROM alpine:3.17

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
    copyright='Copyright (c) 2018 - 2023: Intel'

WORKDIR /
COPY --from=builder /device-scale/device-scale /device-scale
COPY --from=builder /device-scale/cmd/res/configuration.yaml /res/configuration.yaml
COPY --from=builder /device-scale/cmd/res/profiles/cas-scale.yaml /res/profiles/cas-scale.yaml
COPY --from=builder /device-scale/cmd/res/devices/device-list.yaml /res/devices/device-list.yaml

CMD [ "/device-scale","--cp=consul://edgex-core-consul:8500"]
