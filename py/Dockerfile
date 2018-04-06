FROM thrift:latest as builder
ADD *.thrift /thrift/
WORKDIR /thrift
RUN thrift -r -gen py transaction.thrift
RUN thrift -r -gen py database.thrift
RUN thrift -r -gen py cache.thrift
RUN thrift -r -gen py logger.thrift
RUN thrift -r -gen py locker.thrift
RUN thrift -r -gen py distqueue.thrift

FROM python:latest
ADD . /src
WORKDIR /src
RUN pip install -r requirements.txt
COPY --from=builder /thrift/gen-py /src/gen-py
RUN python --version