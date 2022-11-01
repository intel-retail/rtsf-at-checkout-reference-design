# Copyright © 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

# FROM golang:1.13-alpine AS builder
FROM golang:1.18-alpine3.16 AS builder

ENV GO111MODULE=on
WORKDIR /device-scale

LABEL license='SPDX-License-Identifier: Apache-2.0' \
    copyright='Copyright (c) 2018 - 2020: Intel'

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

# add git for go modules
RUN apk update && apk add --no-cache make git gcc libc-dev zeromq-dev linux-headers

COPY go.mod .

RUN go mod download

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -o device-scale -ldflags='-s -w' -a -installsuffix cgo cmd/main.go

# Next image - Copy built Go binary into new workspace
FROM golang:1.16-alpine

LABEL license='SPDX-License-Identifier: Apache-2.0' \
    copyright='Copyright (c) 2018 - 2022: Intel'

WORKDIR /
COPY --from=builder /device-scale/device-scale /device-scale
COPY --from=builder /device-scale/cmd/res/configuration.toml /res/configuration.toml
COPY --from=builder /device-scale/cmd/res/device-scale.yaml /res/device-scale.yaml

CMD [ "/device-scale","--cp=consul://edgex-core-consul:8500","--registry","--confdir=/res"]
