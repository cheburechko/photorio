FROM anibali/pytorch:2.0.1-cuda11.8

USER root

WORKDIR /root

RUN pip3 install transformers requests click kafka-python

COPY "model/*.json" ./model/
COPY "model/*.txt" ./model/
COPY "model/pytorch_model.bin" ./model/

COPY "python/*.py" .
COPY "build/config.yaml" .

CMD ["python3", "__init__.py", "--config-path", "./config.yaml", "run"]
