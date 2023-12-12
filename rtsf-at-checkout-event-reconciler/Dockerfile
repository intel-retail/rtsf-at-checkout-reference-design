# Copyright © 2019 - 2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM golang:1.21-alpine3.18 AS builder

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
    copyright='Copyright (c) 2019 - 2023: Intel'

RUN apk update && apk add --no-cache make git
ENV GO111MODULE=on
WORKDIR /event-reconciler

COPY go.mod .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o event-reconciler /event-reconciler/main.go

# Next image - Copy built Go binary into new workspace
FROM alpine:3.18

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
    copyright='Copyright (c) 2019 - 2023: Intel'

COPY --from=builder /event-reconciler/res /res
COPY --from=builder /event-reconciler/event-reconciler /event-reconciler


CMD [ "/event-reconciler" ,"--cp=consul://edgex-core-consul:8500","--registry"]

