import requests
import click
from PIL import Image
from transformers import BlipProcessor, BlipForConditionalGeneration


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


@click.command()
@click.option('--model-path', default="./blip-image-captioning-large", help='Path to model')
@click.option('--image-url', default='https://storage.googleapis.com/sfr-vision-language-research/BLIP/demo.jpg', help='Image url')
def caption(model_path, image_url):
    """Caption the image at the url"""
    raw_image = fetch_image(image_url)

    model = Model(model_path)

    result = model.caption(raw_image)

    print(result)

if __name__ == '__main__':
    caption()
