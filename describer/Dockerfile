FROM anibali/pytorch:2.0.1-cuda11.8

USER root

WORKDIR /root

COPY *.py .
COPY config.yaml .
COPY "./blip-image-captioning-large/*.json" ./blip-image-captioning-large/
COPY "./blip-image-captioning-large/*.txt" ./blip-image-captioning-large/
COPY "./blip-image-captioning-large/pytorch_model.bin" ./blip-image-captioning-large/

RUN pip3 install transformers requests click kafka-python

CMD ["python3", "__init__.py", "--config-path", "./config.yaml", "run"]
