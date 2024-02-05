from PIL import Image
from transformers import BlipProcessor, BlipForConditionalGeneration


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
