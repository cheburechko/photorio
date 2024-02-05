from PIL import Image
import requests


def fetch_image(url: str) -> Image.Image:
    return Image.open(requests.get(url, stream=True).raw).convert('RGB')
