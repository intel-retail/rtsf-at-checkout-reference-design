# Copyright © 2022 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

FROM python:3.8-bullseye

RUN apt-get update && apt-get install --no-install-recommends libsm6=2:1.2.3-1 libxext6=2:1.3.3-1.1 -y \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY ./requirements.txt /requirements.txt
RUN pip3 install --upgrade pip --no-cache-dir -r /requirements.txt

COPY  ./src /cv-region-of-interest

WORKDIR /cv-region-of-interest

CMD ["python3","-u","enterexit.py"]