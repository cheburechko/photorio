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
            **config,
        )

    def consume(self, model, es):
        for msg in self.consumer:
            try:
                image_url = json.loads(msg.value.decode('ascii')).get("url")
                raw_image = fetch.fetch_image(image_url)
                caption = model.caption(raw_image)
                es.index(image_url, caption)
            except Exception:
                logger.exception("Failed to process message")
