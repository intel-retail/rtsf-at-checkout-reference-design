
FROM python:3-alpine

RUN pip3 install paho-mqtt requests
COPY  ./src /cv-region-of-interest

WORKDIR /cv-region-of-interest

CMD ["python3","-u","enterexit.py"]