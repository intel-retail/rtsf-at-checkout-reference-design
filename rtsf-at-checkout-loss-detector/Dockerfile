# Copyright © 2019 - 2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM golang:1.20-alpine3.17 AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
    copyright='Copyright (c) 2019 - 2023: Intel'

# add git for go modules
RUN apk update && apk add --no-cache make git
ENV GO111MODULE=on
WORKDIR /loss-detector

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o loss-detector /loss-detector/main.go

# Next image - Copy built Go binary into new workspace
FROM alpine:3.17

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
    copyright='Copyright (c) 2019 - 2023: Intel'

COPY --from=builder /loss-detector/res/ /res/
COPY --from=builder /loss-detector/loss-detector /loss-detector

CMD [ "/loss-detector", "--cp=consul://edgex-core-consul:8500", "--registry"]
