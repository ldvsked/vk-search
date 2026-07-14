#!/usr/bin/env bash
set -e

echo "Waiting for PostgreSQL..."

until python - <<'PY'
import os
import psycopg2

try:
    conn = psycopg2.connect(
        host=os.getenv("POSTGRES_HOST", "postgres_db"),
        port=int(os.getenv("POSTGRES_PORT", "5432")),
        dbname=os.getenv("POSTGRES_DB", "vk_search"),
        user=os.getenv("POSTGRES_USER", "vk_user"),
        password=os.getenv("POSTGRES_PASSWORD", "vk_password"),
    )
    conn.close()
except Exception:
    raise SystemExit(1)
PY
do
  echo "PostgreSQL is not ready yet..."
  sleep 2
done

echo "Waiting for OpenSearch..."

until python - <<'PY'
import os
import requests

url = os.getenv("OPENSEARCH_URL", "http://opensearch:9200")

try:
    response = requests.get(url, timeout=3)
    response.raise_for_status()
except Exception:
    raise SystemExit(1)
PY
do
  echo "OpenSearch is not ready yet..."
  sleep 5
done

echo "Loading seed documents to PostgreSQL..."

for file in data/seed/*.jsonl; do
  echo "Loading $file"
  python parser/03_load_to_postgres.py --input "$file"
done

echo "Creating chunks..."
python parser/04_chunk_documents.py --recreate

echo "Creating OpenSearch index..."
python parser/05_create_opensearch_index.py --recreate

echo "Indexing chunks to OpenSearch..."
python parser/06_index_to_opensearch.py

echo "Seed bootstrap finished."