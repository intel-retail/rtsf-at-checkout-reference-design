# Copyright © 2019 - 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM golang:1.13-alpine AS builder

LABEL license='SPDX-License-Identifier: Apache-2.0' \
    copyright='Copyright (c) 2019 - 2020: Intel'

# add git for go modules
RUN apk update && apk add --no-cache make git gcc libc-dev libsodium-dev zeromq-dev
ENV GO111MODULE=on
WORKDIR /loss-detector

COPY go.mod .

RUN go mod download

COPY . .
RUN apk info -a zeromq-dev

RUN CGO_ENABLED=1 go build -a -installsuffix cgo -o loss-detector /loss-detector/main.go

# Next image - Copy built Go binary into new workspace
FROM alpine:3.10

LABEL license='SPDX-License-Identifier: Apache-2.0' \
    copyright='Copyright (c) 2019 - 2020: Intel'

RUN apk update && apk --no-cache add zeromq

COPY --from=builder /loss-detector/res/ /res/
COPY --from=builder /loss-detector/loss-detector /loss-detector

CMD [ "/loss-detector", "--cp=consul://edgex-core-consul:8500", "--registry", "--confdir=/res"]
