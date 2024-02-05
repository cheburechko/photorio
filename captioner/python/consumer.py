import json
import logging
import os

from kafka import KafkaConsumer

import fetch


logger = logging.getLogger(__name__)


class Consumer:
    def __init__(self, topics, config):
        self.consumer = KafkaConsumer(
            *topics,
            sasl_plain_password=os.getenv("KAFKA_PASSWORD"),
            value_deserializer=lambda m: json.loads(m.decode('ascii')),
            **config,
        )

    def consume(self, model):
        for msg in self.consumer:
            try:
                image_url = msg.value.get("url")
                raw_image = fetch.fetch_image(image_url)
                result = model.caption(raw_image)
                logger.info("URL: %s, caption: %s", image_url, result)
            except Exception:
                logger.exception("Failed to process message")
