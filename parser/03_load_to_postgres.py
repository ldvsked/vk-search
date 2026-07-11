import argparse
import json
from pathlib import Path
from typing import Any

import psycopg2
from psycopg2.extras import Json

from config import get_settings


def read_jsonl(path: Path) -> list[dict[str, Any]]:
    if not path.exists():
        raise FileNotFoundError(f"File not found: {path}")

    rows = []

    with path.open("r", encoding="utf-8") as file:
        for line_number, line in enumerate(file, start=1):
            line = line.strip()

            if not line:
                continue

            try:
                rows.append(json.loads(line))
            except json.JSONDecodeError as exc:
                print(f"Skip invalid JSON line {line_number}: {exc}")

    return rows


def get_connection():
    settings = get_settings()

    return psycopg2.connect(
        host=settings.postgres_host,
        port=settings.postgres_port,
        dbname=settings.postgres_db,
        user=settings.postgres_user,
        password=settings.postgres_password,
    )


def get_or_create_source(
    cursor,
    source_name: str,
    source_domain: str,
) -> int:
    source_url = f"https://vk.com/{source_domain}"

    cursor.execute(
        """
        INSERT INTO sources (name, domain, type, url)
        VALUES (%s, %s, %s, %s)
        ON CONFLICT (domain)
        DO UPDATE SET
            name = EXCLUDED.name,
            url = EXCLUDED.url
        RETURNING id;
        """,
        (
            source_name,
            source_domain,
            "vk_public",
            source_url,
        ),
    )

    source_id = cursor.fetchone()[0]
    return source_id


def insert_document(
    cursor,
    source_id: int,
    row: dict[str, Any],
) -> int:
    cursor.execute(
        """
        INSERT INTO documents (
            source_id,
            external_id,
            title,
            text,
            content_hash,
            url,
            published_at,
            raw_json
        )
        VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
        ON CONFLICT (external_id)
        DO UPDATE SET
            source_id = EXCLUDED.source_id,
            title = EXCLUDED.title,
            text = EXCLUDED.text,
            content_hash = EXCLUDED.content_hash,
            url = EXCLUDED.url,
            published_at = EXCLUDED.published_at,
            raw_json = EXCLUDED.raw_json
        RETURNING id;
        """,
        (
            source_id,
            row["external_id"],
            row.get("title"),
            row["text"],
            row["content_hash"],
            row["url"],
            row.get("published_at"),
            Json(row.get("raw_json", {})),
        ),
    )

    document_id = cursor.fetchone()[0]
    return document_id


def print_db_stats(cursor) -> None:
    cursor.execute("SELECT COUNT(*) FROM sources;")
    sources_count = cursor.fetchone()[0]

    cursor.execute("SELECT COUNT(*) FROM documents;")
    documents_count = cursor.fetchone()[0]

    cursor.execute(
        """
        SELECT
            s.domain,
            COUNT(d.id) AS documents_count
        FROM sources s
        LEFT JOIN documents d ON d.source_id = s.id
        GROUP BY s.domain
        ORDER BY documents_count DESC;
        """
    )
    by_source = cursor.fetchall()

    print("\nDatabase stats:")
    print(f"Sources: {sources_count}")
    print(f"Documents: {documents_count}")

    print("\nDocuments by source:")
    for domain, count in by_source:
        print(f"- {domain}: {count}")


def main() -> None:
    parser = argparse.ArgumentParser(description="Load clean VK posts to PostgreSQL")
    parser.add_argument(
        "--input",
        type=str,
        default="data/clean_posts.jsonl",
        help="Path to clean posts JSONL",
    )
    args = parser.parse_args()

    input_path = Path(args.input)
    rows = read_jsonl(input_path)

    if not rows:
        print("No rows to load")
        return

    inserted_or_updated = 0
    errors = 0
    source_cache: dict[str, int] = {}

    connection = get_connection()

    try:
        with connection:
            with connection.cursor() as cursor:
                for row in rows:
                    try:
                        source_name = row.get("source_name") or "unknown"
                        source_domain = row.get("source_domain") or source_name

                        if source_domain not in source_cache:
                            source_cache[source_domain] = get_or_create_source(
                                cursor=cursor,
                                source_name=source_name,
                                source_domain=source_domain,
                            )

                        source_id = source_cache[source_domain]

                        insert_document(
                            cursor=cursor,
                            source_id=source_id,
                            row=row,
                        )

                        inserted_or_updated += 1

                    except Exception as exc:
                        errors += 1
                        print(
                            f"Skip document {row.get('external_id')}: {exc}"
                        )

                print_db_stats(cursor)

    finally:
        connection.close()

    print("\nLoading finished")
    print(f"Input rows: {len(rows)}")
    print(f"Inserted or updated: {inserted_or_updated}")
    print(f"Errors: {errors}")


if __name__ == "__main__":
    main()