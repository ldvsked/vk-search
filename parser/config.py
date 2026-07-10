import os
from dataclasses import dataclass
from dotenv import load_dotenv


load_dotenv()


@dataclass(frozen=True)
class Settings:
    vk_token: str
    vk_api_version: str
    vk_publics: list[str]
    vk_target_posts: int
    vk_request_sleep_seconds: float

    postgres_host: str
    postgres_port: int
    postgres_db: str
    postgres_user: str
    postgres_password: str

    opensearch_url: str
    opensearch_index: str


def get_settings() -> Settings:
    publics_raw = os.getenv("VK_PUBLICS", "")
    publics = [item.strip() for item in publics_raw.split(",") if item.strip()]

    return Settings(
        vk_token=os.getenv("VK_TOKEN", ""),
        vk_api_version=os.getenv("VK_API_VERSION", "5.199"),
        vk_publics=publics,
        vk_target_posts=int(os.getenv("VK_TARGET_POSTS", "500")),
        vk_request_sleep_seconds=float(os.getenv("VK_REQUEST_SLEEP_SECONDS", "0.35")),

        postgres_host=os.getenv("POSTGRES_HOST", "localhost"),
        postgres_port=int(os.getenv("POSTGRES_PORT", "5432")),
        postgres_db=os.getenv("POSTGRES_DB", "vk_search"),
        postgres_user=os.getenv("POSTGRES_USER", "vk_user"),
        postgres_password=os.getenv("POSTGRES_PASSWORD", "vk_password"),

        opensearch_url=os.getenv("OPENSEARCH_URL", "http://localhost:9200"),
        opensearch_index=os.getenv("OPENSEARCH_INDEX", "vk_chunks"),
    )