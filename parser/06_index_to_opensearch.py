import argparse
from datetime import datetime, timezone
from typing import Any

import psycopg2
from opensearchpy import OpenSearch
from opensearchpy.helpers import bulk

from config import get_settings


def get_postgres_connection():
    settings = get_settings()

    return psycopg2.connect(
        host=settings.postgres_host,
        port=settings.postgres_port,
        dbname=settings.postgres_db,
        user=settings.postgres_user,
        password=settings.postgres_password,
    )


def get_opensearch_client() -> OpenSearch:
    settings = get_settings()

    return OpenSearch(
        hosts=[settings.opensearch_url],
        verify_certs=False,
        ssl_show_warn=False,
    )


def fetch_chunks(cursor) -> list[dict[str, Any]]:
    cursor.execute(
        """
        SELECT
            c.id AS chunk_id,
            c.document_id,
            d.source_id,
            s.name AS source_name,
            d.title,
            c.content,
            d.url,
            d.published_at
        FROM chunks c
        JOIN documents d ON d.id = c.document_id
        JOIN sources s ON s.id = d.source_id
        ORDER BY c.id;
        """
    )

    rows = cursor.fetchall()

    result = []

    for row in rows:
        published_at = row[7]

        result.append(
            {
                "chunk_id": row[0],
                "document_id": row[1],
                "source_id": row[2],
                "source_name": row[3],
                "title": row[4],
                "content": row[5],
                "url": row[6],
                "published_at": published_at.isoformat() if published_at else None,
            }
        )

    return result


def build_bulk_actions(
    chunks: list[dict[str, Any]],
    index_name: str,
) -> list[dict[str, Any]]:
    actions = []

    for chunk in chunks:
        opensearch_id = str(chunk["chunk_id"])

        actions.append(
            {
                "_index": index_name,
                "_id": opensearch_id,
                "_source": chunk,
            }
        )

    return actions


def update_chunks_opensearch_ids(cursor, chunks: list[dict[str, Any]]) -> None:
    for chunk in chunks:
        cursor.execute(
            """
            UPDATE chunks
            SET opensearch_id = %s
            WHERE id = %s;
            """,
            (
                str(chunk["chunk_id"]),
                chunk["chunk_id"],
            ),
        )


def create_index_run_start(cursor) -> int:
    cursor.execute(
        """
        INSERT INTO index_runs (
            status,
            started_at
        )
        VALUES (%s, %s)
        RETURNING id;
        """,
        (
            "running",
            datetime.now(timezone.utc),
        ),
    )

    return cursor.fetchone()[0]


def finish_index_run(
    cursor,
    run_id: int,
    status: str,
    documents_count: int,
    chunks_count: int,
    error_message: str | None = None,
) -> None:
    cursor.execute(
        """
        UPDATE index_runs
        SET
            status = %s,
            finished_at = %s,
            documents_count = %s,
            chunks_count = %s,
            error_message = %s
        WHERE id = %s;
        """,
        (
            status,
            datetime.now(timezone.utc),
            documents_count,
            chunks_count,
            error_message,
            run_id,
        ),
    )


def print_opensearch_stats(client: OpenSearch, index_name: str) -> None:
    client.indices.refresh(index=index_name)

    count_response = client.count(index=index_name)
    count = count_response.get("count", 0)

    print("\nOpenSearch stats:")
    print(f"Index: {index_name}")
    print(f"Documents in index: {count}")


def main() -> None:
    parser = argparse.ArgumentParser(description="Index PostgreSQL chunks to OpenSearch")
    parser.add_argument(
        "--batch-size",
        type=int,
        default=500,
        help="Bulk indexing batch size",
    )
    args = parser.parse_args()

    settings = get_settings()
    index_name = settings.opensearch_index

    pg_connection = get_postgres_connection()
    os_client = get_opensearch_client()

    run_id = None

    try:
        with pg_connection:
            with pg_connection.cursor() as cursor:
                run_id = create_index_run_start(cursor)

                chunks = fetch_chunks(cursor)

                if not chunks:
                    print("No chunks found. Run parser/04_chunk_documents.py first.")
                    finish_index_run(
                        cursor=cursor,
                        run_id=run_id,
                        status="failed",
                        documents_count=0,
                        chunks_count=0,
                        error_message="No chunks found",
                    )
                    return

                actions = build_bulk_actions(
                    chunks=chunks,
                    index_name=index_name,
                )

                success_count, errors = bulk(
                    client=os_client,
                    actions=actions,
                    chunk_size=args.batch_size,
                    raise_on_error=False,
                )

                update_chunks_opensearch_ids(cursor, chunks)

                cursor.execute("SELECT COUNT(*) FROM documents;")
                documents_count = cursor.fetchone()[0]

                finish_index_run(
                    cursor=cursor,
                    run_id=run_id,
                    status="success" if not errors else "partial_success",
                    documents_count=documents_count,
                    chunks_count=len(chunks),
                    error_message=str(errors[:5]) if errors else None,
                )

                print("\nIndexing finished")
                print(f"Chunks read from PostgreSQL: {len(chunks)}")
                print(f"Successfully indexed: {success_count}")
                print(f"Errors: {len(errors) if errors else 0}")

                if errors:
                    print("First errors:")
                    for error in errors[:5]:
                        print(error)

        print_opensearch_stats(os_client, index_name=index_name)

    except Exception as exc:
        if run_id is not None:
            with pg_connection:
                with pg_connection.cursor() as cursor:
                    finish_index_run(
                        cursor=cursor,
                        run_id=run_id,
                        status="failed",
                        documents_count=0,
                        chunks_count=0,
                        error_message=str(exc),
                    )

        raise

    finally:
        pg_connection.close()


if __name__ == "__main__":
    main()