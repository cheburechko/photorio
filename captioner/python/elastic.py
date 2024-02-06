import hashlib
import logging

from elasticsearch import Elasticsearch

logger = logging.getLogger(__name__)


class ElasticClient:
    def __init__(self, addr, index) -> None:
        self.client = Elasticsearch(addr)
        self.index_name = index
    
    def index(self, image_url, caption):
        logger.info("URL: %s, caption: %s", image_url, caption)
        self.client.index(
            index=self.index_name,
            id=hashlib.md5(image_url.encode(), usedforsecurity=False).hexdigest(),
            document={
                "url": image_url,
                "caption": caption,
            }
        )
