import io
import logging
import os

import kafka
import numpy as np
import pytest
from PIL import Image

from consumer import Consumer


class StubModel:
    def caption(self, raw_image: Image.Image) -> str:
        logging.info(repr(raw_image))
        return repr(raw_image)


class StubES:
    def index(self, image_url, caption):
        logging.info("URL: %s, caption: %s", image_url, caption)


@pytest.fixture
def kafka_config():
    yield {
        "bootstrap_servers": ['localhost:9092'],
        "security_protocol": "SASL_PLAINTEXT",
        "sasl_mechanism": "PLAIN",
        "sasl_plain_username": "user1",
        "sasl_plain_password": os.getenv("KAFKA_PASSWORD"),
    }


@pytest.fixture
def kafka_consumer_config(kafka_config):
    return dict(kafka_config, group_id="test")


@pytest.fixture
def topic(kafka_config):
    admin_client = kafka.KafkaAdminClient(**kafka_config)
    result = "test"

    try:
        admin_client.delete_topics([result])
    except:
        pass
    
    admin_client.create_topics([result])
    topic_config = kafka.ConfigResource(restype='TOPIC', name=result, set_config={"max.message.bytes":f"{2**20 * 10}"}, described_configs=None, error=None)
    admin_client.alter_configs([topic_config])
    yield result
    admin_client.delete_topics([result])


def test_consume(kafka_config, kafka_consumer_config, topic):
    producer = kafka.KafkaProducer(**kafka_config)

    img = Image.fromarray(np.random.randint(0, 255, (1024,1024), dtype=np.dtype('uint8')))
    img_byte_arr = io.BytesIO()
    img.save(img_byte_arr, format='PNG')

    for _ in range(10):
        future = producer.send(topic, img_byte_arr.getvalue())
        future.get(timeout=60)

    consumer = Consumer([topic], kafka_consumer_config)
    consumer.consume(StubModel(), StubES())
