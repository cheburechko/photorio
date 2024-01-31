FROM anibali/pytorch:2.0.1-cuda11.8

USER root

WORKDIR /root

COPY *.py .
COPY "./blip-image-captioning-large/*.json" ./model/
COPY "./blip-image-captioning-large/*.txt" ./model/
COPY "./blip-image-captioning-large/pytorch_model.bin" ./model/

RUN pip3 install transformers requests click

CMD ["python3", "__init__.py", "--model-path", "./model"]
