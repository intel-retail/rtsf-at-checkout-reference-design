# Copyright © 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM python:3.11-alpine

COPY ./requirements.txt /requirements.txt
RUN pip3 install --upgrade pip --no-cache-dir -r /requirements.txt

COPY  ./src /cv-region-of-interest

WORKDIR /cv-region-of-interest

CMD ["python3","-u","enterexit.py"]