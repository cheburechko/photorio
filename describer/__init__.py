from dataclasses import dataclass
import json
import logging
import os
import sys

import click
from kafka import KafkaConsumer
from PIL import Image
import requests
from transformers import BlipProcessor, BlipForConditionalGeneration
import yaml


logger = logging.getLogger(__name__)


class BaseModel:
    def caption(self, raw_image: Image.Image) -> str:
        raise NotImplemented


class BlipModel(BaseModel):
    def __init__(self, path: str) -> None:
        self.processor = BlipProcessor.from_pretrained(path)
        self.model = BlipForConditionalGeneration.from_pretrained(path).to("cuda")
    
    def caption(self, raw_image: Image.Image) -> str:
        inputs = self.processor(raw_image, return_tensors="pt").to("cuda")
        out = self.model.generate(max_new_tokens=20, **inputs)
        return self.processor.decode(out[0], skip_special_tokens=True)


class StubModel(BaseModel):
    def __init__(self, path: str) -> None:
        pass
    
    def caption(self, raw_image: Image.Image) -> str:
        return repr(raw_image)


def fetch_image(url: str) -> Image.Image:
    return Image.open(requests.get(url, stream=True).raw).convert('RGB')


def setup_logging(logger):
    logger.setLevel(logging.INFO)

    handler = logging.StreamHandler()
    handler.setFormatter(logging.Formatter('%(asctime)s - %(name)s - %(funcName)s - %(levelname)s - %(message)s'))
    logger.addHandler(handler)


@dataclass
class Context:
    config: dict
    model: BaseModel


@click.group()
@click.option('--config-path', default='./config.yaml', help='Config path')
@click.pass_context
def cli(ctx, config_path):
    setup_logging(logger)
    with open(config_path) as f:
        config = yaml.safe_load(f)

    ctx.ensure_object(dict)
    model_cls = StubModel if config["debug"] else BlipModel
    ctx.obj = Context(config, model_cls(config["model_path"]))


@cli.command()
@click.option('--image-url', default='https://storage.googleapis.com/sfr-vision-language-research/BLIP/demo.jpg', help='Image url')
@click.pass_context
def caption(ctx, image_url):
    """Caption the image at the url"""
    raw_image = fetch_image(image_url)

    result = ctx.obj.model.caption(raw_image)

    print(result)


@cli.command()
@click.pass_context
def run(ctx):
    """Read messages from Kafka and describe urls"""
    logger.info("Starting")

    consumer = KafkaConsumer(
        *ctx.obj.config["topics"],
        sasl_plain_password=os.getenv("KAFKA_PASSWORD"),
        value_deserializer=lambda m: json.loads(m.decode('ascii')),
        **ctx.obj.config["kafka_config"],
    )

    logger.info("Initialized consumer")

    for msg in consumer:
        try:
            image_url = msg.value.get("url")
            raw_image = fetch_image(image_url)
            result = ctx.obj.model.caption(raw_image)
            logger.info("URL: %s, caption: %s", image_url, result)
        except Exception:
            logger.exception("Failed to process message")
    
    logger.info("Finished consuming")

    if consumer is not None:
        consumer.close()


if __name__ == '__main__':
    cli()
