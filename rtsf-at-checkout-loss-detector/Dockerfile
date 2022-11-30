# Copyright © 2019 - 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM golang:1.18-alpine3.16 AS builder

LABEL license='SPDX-License-Identifier: Apache-2.0' \
    copyright='Copyright (c) 2019 - 2022: Intel'

# add git for go modules
RUN apk update && apk add --no-cache make git
ENV GO111MODULE=on
WORKDIR /loss-detector

COPY go.mod .

RUN go mod download

COPY . .

RUN go build -tags no_zmq -o loss-detector /loss-detector/main.go

# Next image - Copy built Go binary into new workspace
FROM alpine:3.16

LABEL license='SPDX-License-Identifier: Apache-2.0' \
    copyright='Copyright (c) 2019 - 2022: Intel'

COPY --from=builder /loss-detector/res/ /res/
COPY --from=builder /loss-detector/loss-detector /loss-detector

CMD [ "/loss-detector", "--cp=consul://edgex-core-consul:8500", "--registry", "--confdir=/res"]
