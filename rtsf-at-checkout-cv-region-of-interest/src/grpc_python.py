#
# Copyright (C) 2024 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

import os
import argparse
import datetime
import cv2
import grpc
import numpy as np
import json
import time
from tensorflow import make_tensor_proto, make_ndarray
from tensorflow_serving.apis import predict_pb2
from tensorflow_serving.apis import prediction_service_pb2_grpc
import paho.mqtt.publish as publish
import sys
sys.path.append("/model_server/demos/common/python")


def openInputSrc(input_src):
    # OpenCV RTSP Stream
    stream = cv2.VideoCapture(input_src)
    if not stream.isOpened():
        print('Unable to open source:' + input_src)
        exit(-1)
    return stream


def setupGRPC(address, port):
    address = "{}:{}".format(address, port)
    # Create gRPC stub for communicating with the server
    channel = grpc.insecure_channel(address)
    grpc_stub = prediction_service_pb2_grpc.PredictionServiceStub(channel)

    return grpc_stub


def getModelInputImageSize(model_name):
    if model_name == "instance-segmentation-security-1040":
        return [608, 608]
    elif model_name == "bit_64":
        return [64, 64]
    elif model_name == "yolov5s":
        return [416, 416]
    elif model_name == "product-detection-0001":
        return [512, 512]
    elif model_name == "ssd_mobilenet_v1_coco":
        return [300, 300]
    else:
        return None


def getInputName(model_name):
    if model_name == "instance-segmentation-security-1040":
        return "image"
    elif model_name == "bit_64" or model_name == "product-detection-0001":
        return "input_1"
    elif model_name == "yolov5s":
        return "images"
    elif model_name == "ssd_mobilenet_v1_coco":
        return "image_tensor"
    else:
        return None


def getOutputName(model_name):
    if model_name == "instance-segmentation-security-1040":
        return "mask"
    elif model_name == "bit_64":
        return "output_1"
    elif model_name == "yolov5s":
        return "326/sink_port_0"
    else:
        return None


def inference(img_str, model_name, grpc_stub):
    request = predict_pb2.PredictRequest()
    request.model_spec.name = model_name
    request.inputs['input.1'].CopyFrom(
        make_tensor_proto(img_str, shape=img_str.shape)
    )
    start_time = datetime.datetime.now()
    response = None
    try:
        response = grpc_stub.Predict(request, 30.0)
    except grpc.RpcError as err:
        print("Encountered gRPC error")
        if err.code() == grpc.StatusCode.ABORTED:
            print('No product has been found in the image')
            exit(1)
        else:
            raise err

    end_time = datetime.datetime.now()
    duration = (end_time - start_time).total_seconds() * 1000
    return [response, duration]


def create_inference_pipeline(
        input_src, roi_name, mqtt_broker_address, mqtt_outgoing_topic,
        grpc_address='localhost',
        grpc_port=9001, model_name='product-detection-0001'):
    '''
    parser = argparse.ArgumentParser(description='Sends requests
                                    via KServe gRPC API using images in format
                                    supported by OpenCV. It displays
                                    performance statistics and optionally the
                                    model accuracy')
    parser.add_argument('--input_src', required=True, default='',
                        help='input source for the inference pipeline')
    parser.add_argument('--grpc_address',required=False, default='localhost',
                        help='Specify url to grpc service. default:localhost')
    parser.add_argument('--grpc_port',required=False, default=9000,
                        help='Specify port to grpc service. default: 9000')
    parser.add_argument('--model_name',
                        default='instance-segmentation-security-1040',
                        help='Define model name,
                        must be same as is in service. default: resnet',
                        dest='model_name')
    args = vars(parser.parse_args())'''

    print("Connect to stream")
    stream = openInputSrc(input_src)

    print("Establish OVMS GRPc connection")
    grpc_stub = setupGRPC(grpc_address, grpc_port)

    print("Get the model size from OVMS metadata")
    model_input_image_size = getModelInputImageSize(model_name)
    model_name = model_name
    print("model_name")
    print(model_name)
    print("model_input_image_size")
    print(model_input_image_size)
    print("Begin inference loop")
    frame_id = 0
    while True:
        try:
            # get frame from OpenCV
            _, frame = stream.read()

            # image pre-processing
            img = frame.astype(np.float32)
            resize_to_shape = getModelInputImageSize(model_name)
            img = cv2.resize(
                img, (resize_to_shape[1], resize_to_shape[0])
            )
            img = img.transpose(2, 0, 1).reshape(
                1, 3, resize_to_shape[0], resize_to_shape[1]
            )

            response = inference(img, model_name, grpc_stub)

            if model_name == "instance-segmentation-security-1040":
                postProcessMaskRCNN(response[0], response[1])
            elif model_name == "bit_64":
                postProcessBit(response[0], response[1])
            elif model_name == "yolov5s":
                postProcessYolov5s(response[0], response[1])
            elif model_name == "product-detection-0001":
                detected_products_jsonMsg = post_process_product_detection(
                    input_src, frame_id, roi_name, response[0], response[1],
                    img, resize_to_shape[0], resize_to_shape[1],
                    mqtt_broker_address, mqtt_outgoing_topic
                )
                publish_mqtt_msg(
                    detected_products_jsonMsg, mqtt_broker_address,
                    mqtt_outgoing_topic
                )
            else:
                print("Unsupported model_name: {}".format(model_name))
                exit(1)
            frame_id = frame_id + 1
        except Exception as e:
            print(e)
            pass  # nosec


def post_process_product_detection(
        input_src, frame_id, roi_name, result, duration, img, width, height,
        mqtt_broker_address, mqtt_outgoing_topic):
    products = []
    for name in result.outputs:
        print(f"Output: name[{name}]")

        output = make_ndarray(result.outputs["868"])
        print("Response shape", output.shape)
        print("img.shape[0]: width" + str(width))
        print("img.shape[1]: height" + str(height))

        # iterate over responses from all images in the batch
        for y in range(0, img.shape[0]):
            img_out = img[y, :, :, :]

            # print("image in batch item",y, ", output shape",img_out.shape)
            img_out = img_out.transpose(1, 2, 0)
            # there is returned 200 detections for each image in the batch
            for i in range(0, 200*1-1):
                detection = output[:, :, i, :]
                label_id = 0
                # each detection has shape 1,1,7
                # where last dimension represent:
                # image_id - ID of the image in the batch
                # label - predicted class ID
                # conf - confidence for the predicted class
                # (x_min,y_min)-coordinates of top left bounding box corner
                # (x_max,y_max)-coordinates of bottom right bounding box corner
                # ignore detections for image_id != y and confidence <0.5
                if detection[0, 0, 2] > 0.5 and int(detection[0, 0, 0]) == y:
                    print("detection", i, detection)
                    product_number = detection[0, 0, 1]
                    x_min = int(detection[0, 0, 3] * width)
                    y_min = int(detection[0, 0, 4] * height)
                    x_max = int(detection[0, 0, 5] * width)
                    y_max = int(detection[0, 0, 6] * height)
                    # box coordinates are proportional to the image size
                    # print("product_number", product_number)
                    # print("x_min", x_min)
                    # print("y_min", y_min)
                    # print("x_max", x_max)
                    # print("y_max", y_max)

                    # model specific labels
                    detection_str = {
                        0: "background_label",
                        1: "undefined",
                        2: "sprite",
                        3: "kool-aid",
                        4: "extra",
                        5: "ocelo",
                        6: "finish",
                        7: "mtn_dew",
                        8: "best_foods",
                        9: "gatorade",
                        10: "heinz",
                        11: "ruffles",
                        12: "pringles",
                        13: "del_monte"
                    }[product_number]
                    detected_product = {"product": detection_str,
                                        "confidence": str(detection[0, 0, 2]),
                                        "x_min": str(x_min),
                                        "y_min": str(y_min),
                                        "x_max": str(x_max),
                                        "y_max": str(y_max)}
                    if len(products) < i + 1:
                        products.append(detected_product)
                    else:
                        products[i].update(detected_product)
    detected_products = {"timestamp": str(time.time()),
                         "source": input_src,
                         "roi_name": roi_name,
                         "frame_id": frame_id,
                         "width": str(width),
                         "height": str(height),
                         "objects": products}
    detected_products_jsonMsg = json.dumps(detected_products)
    return detected_products_jsonMsg


def publish_mqtt_msg(
        detected_products, mqtt_broker_address, mqtt_outgoing_topic):
    publish.single(
        mqtt_outgoing_topic, detected_products,  hostname=mqtt_broker_address
    )


if __name__ == '__main__':
    camSrc = 'file:///home/nesubuntu207/Neethu/rtsf/NES_fork/\
    rtsf-at-checkout-reference-design/loss-detection-app/video-samples/\
    grocery-test.mp4'
    create_inference_pipeline(
        camSrc,
        'Staging', 'localhost', 'AnalyticsData', 'localhost', 9001,
        'product-detection-0001')
