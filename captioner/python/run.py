from dataclasses import dataclass
import logging
import os

import click
import yaml

import elastic
import fetch
import model
from consumer import Consumer


logger = logging.getLogger(__name__)


def setup_logging():
    logger = logging.getLogger()
    logger.setLevel(logging.INFO)

    handler = logging.StreamHandler()
    handler.setFormatter(logging.Formatter('%(asctime)s - %(name)s - %(funcName)s - %(levelname)s - %(message)s'))
    logger.addHandler(handler)

    kafka_logger = logging.getLogger("kafka")
    kafka_logger.setLevel(logging.ERROR)


@dataclass
class Context:
    config: dict
    model: model.BaseModel


@click.group()
@click.option('--config-path', default='../build/config.yaml', help='Config path')
@click.pass_context
def cli(ctx, config_path):
    setup_logging()
    with open(config_path) as f:
        config = yaml.safe_load(f)

    ctx.ensure_object(dict)
    model_cls = model.StubModel if os.getenv("DEBUG") else model.BlipModel
    ctx.obj = Context(config, model_cls(config["model_path"]))


@cli.command()
@click.option('--image-url', default='https://storage.googleapis.com/sfr-vision-language-research/BLIP/demo.jpg', help='Image url')
@click.pass_context
def caption(ctx, image_url):
    """Caption the image at the url"""
    raw_image = fetch.fetch_image(image_url)

    result = ctx.obj.model.caption(raw_image)

    print(result)


@cli.command()
@click.pass_context
def run(ctx):
    """Read messages from Kafka and describe urls"""
    consumer = Consumer(ctx.obj.config["topics"], ctx.obj.config["kafka_config"])
    logger.info("Initialized consumer")

    es = elastic.ElasticClient(ctx.obj.config["elastic_server"], ctx.obj.config["index"])
    logger.info("Initialized elastic")

    consumer.consume(ctx.obj.model, es)
    logger.info("Finished consuming")


if __name__ == '__main__':
    cli()
