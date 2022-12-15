# Copyright © 2019 -2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM golang:1.18-alpine3.16 AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
    copyright='Copyright (c) 2019 - 2022: Intel'

RUN apk update && apk add --no-cache make git
ENV GO111MODULE=on
WORKDIR /event-reconciler

COPY go.mod .
RUN go mod download

COPY . .
RUN go build -tags no_zmq -o event-reconciler /event-reconciler/main.go

# Next image - Copy built Go binary into new workspace
FROM alpine:3.16

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
    copyright='Copyright (c) 2019 -2022: Intel'

COPY --from=builder /event-reconciler/res /res
COPY --from=builder /event-reconciler/event-reconciler /event-reconciler


CMD [ "/event-reconciler" ,"--cp=consul://edgex-core-consul:8500","--registry","--confdir=/res"]

