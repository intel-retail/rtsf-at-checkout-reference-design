# Copyright © 2019 - 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM golang:1.21-alpine3.18 AS builder

ENV GO111MODULE=on
WORKDIR /product-lookup

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
    copyright='Copyright (c) 2018 - 2022: Intel'

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

# add git for go modules
RUN apk update && apk add --no-cache make git

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o product-lookup main.go

# Next image - Copy built Go binary into new workspace
FROM alpine:3.18

LABEL license='SPDX-License-Identifier: BSD-3-Clause' \
    copyright='Copyright (c) 2018 - 2020: Intel'

ENV APP_PORT=8083
EXPOSE $APP_PORT

WORKDIR /
COPY --from=builder /product-lookup/product-lookup /product-lookup
COPY --from=builder /product-lookup/db_initialization/ /db_initialization/

CMD ["/product-lookup", "-file", "/db_initialization/all-products.json"]
