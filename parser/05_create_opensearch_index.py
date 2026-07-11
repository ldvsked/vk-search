import argparse

from opensearchpy import OpenSearch

from config import get_settings


def get_client() -> OpenSearch:
    settings = get_settings()

    return OpenSearch(
        hosts=[settings.opensearch_url],
        verify_certs=False,
        ssl_show_warn=False,
    )


def build_index_body() -> dict:
    return {
        "settings": {
            "index": {
                "number_of_shards": 1,
                "number_of_replicas": 0,
            }
        },
        "mappings": {
            "properties": {
                "chunk_id": {"type": "long"},
                "document_id": {"type": "long"},
                "source_id": {"type": "long"},
                "source_name": {"type": "keyword"},
                "title": {"type": "text"},
                "content": {"type": "text"},
                "url": {"type": "keyword"},
                "published_at": {"type": "date"},
            }
        },
    }


def main() -> None:
    parser = argparse.ArgumentParser(description="Create OpenSearch index")
    parser.add_argument(
        "--recreate",
        action="store_true",
        help="Delete index if it already exists and create it again",
    )
    args = parser.parse_args()

    settings = get_settings()
    client = get_client()
    index_name = settings.opensearch_index

    info = client.info()
    print("Connected to OpenSearch")
    print(f"Cluster name: {info.get('cluster_name')}")

    index_exists = client.indices.exists(index=index_name)

    if index_exists and args.recreate:
        print(f"Deleting existing index: {index_name}")
        client.indices.delete(index=index_name)
        index_exists = False

    if index_exists:
        print(f"Index already exists: {index_name}")
        return

    body = build_index_body()

    client.indices.create(
        index=index_name,
        body=body,
    )

    print(f"Index created: {index_name}")


if __name__ == "__main__":
    main()