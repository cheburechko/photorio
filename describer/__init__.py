from dataclasses import dataclass
import json
from time import sleep

import click
from kafka import KafkaConsumer
from PIL import Image
import requests
from transformers import BlipProcessor, BlipForConditionalGeneration
import yaml


class Model:
    def __init__(self, path: str) -> None:
        self.processor = BlipProcessor.from_pretrained(path)
        self.model = BlipForConditionalGeneration.from_pretrained(path).to("cuda")
    
    def caption(self, raw_image: Image.Image) -> str:
        inputs = self.processor(raw_image, return_tensors="pt").to("cuda")
        out = self.model.generate(max_new_tokens=20, **inputs)
        return self.processor.decode(out[0], skip_special_tokens=True)
    

def fetch_image(url: str) -> Image.Image:
    return Image.open(requests.get(url, stream=True).raw).convert('RGB')


@dataclass
class Context:
    """Class for keeping track of an item in inventory."""
    config: dict
    model: Model


@click.group()
@click.option('--config-path', default='./config.yaml', help='Config path')
@click.pass_context
def cli(ctx, config_path):
    with open(config_path) as f:
        config = yaml.safe_load(f)

    ctx.ensure_object(dict)
    ctx.obj = Context(config, Model(config["model_path"]))


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
    consumer = KafkaConsumer(
        *ctx.obj.config["topics"],
        **ctx.obj.config["kafka_config"],
    )

    for msg in consumer:
        image_url = json.loads(msg)["url"]
        raw_image = fetch_image(image_url)
        result = ctx.obj.model.caption(raw_image)
        print(result)

    if consumer is not None:
        consumer.close()


if __name__ == '__main__':
    cli()
