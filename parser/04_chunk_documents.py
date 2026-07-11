import argparse
import hashlib
import re
from dataclasses import dataclass


import psycopg2

from config import get_settings


@dataclass
class Document:
    id: int
    text: str


def calculate_hash(text: str) -> str:
    return hashlib.sha256(text.encode("utf-8")).hexdigest()


def get_connection():
    settings = get_settings()

    return psycopg2.connect(
        host=settings.postgres_host,
        port=settings.postgres_port,
        dbname=settings.postgres_db,
        user=settings.postgres_user,
        password=settings.postgres_password,
    )


def split_long_paragraph(paragraph: str, max_chars: int) -> list[str]:
    """
    Если один абзац слишком длинный, режем его по предложениям.
    Если предложение тоже огромное, режем грубо по max_chars.
    """
    sentences = re.split(r"(?<=[.!?…])\s+", paragraph)
    chunks = []
    current = ""

    for sentence in sentences:
        sentence = sentence.strip()

        if not sentence:
            continue

        if len(sentence) > max_chars:
            if current:
                chunks.append(current.strip())
                current = ""

            for i in range(0, len(sentence), max_chars):
                part = sentence[i:i + max_chars].strip()
                if part:
                    chunks.append(part)
            continue

        if len(current) + len(sentence) + 1 <= max_chars:
            current = f"{current} {sentence}".strip()
        else:
            if current:
                chunks.append(current.strip())
            current = sentence

    if current:
        chunks.append(current.strip())

    return chunks

def merge_small_chunks(
    chunks: list[str],
    min_chars: int,
    max_chars: int,
) -> list[str]:
    """
    Склеивает слишком маленькие чанки с соседями.
    Если из-за склейки чанк немного превысит max_chars, это допустимо:
    лучше один чанк 1850-2000 символов, чем отдельный мусорный чанк на 18 символов.
    """
    result: list[str] = []

    for chunk in chunks:
        chunk = chunk.strip()

        if not chunk:
            continue

        if len(chunk) < min_chars and result:
            result[-1] = f"{result[-1]}\n\n{chunk}".strip()
        else:
            result.append(chunk)

    if len(result) >= 2 and len(result[0]) < min_chars:
        result[1] = f"{result[0]}\n\n{result[1]}".strip()
        result = result[1:]

    return result

def chunk_text(text: str, max_chars: int = 1800, min_chars: int = 300) -> list[str]:
    """
    Делим текст на осмысленные чанки.

    Логика:
    - если текст короткий, оставляем одним чанком;
    - сначала пытаемся резать по абзацам;
    - если абзац слишком длинный, режем по предложениям;
    - слишком маленькие куски стараемся присоединять к соседним.
    """
    text = text.strip()

    if not text:
        return []

    if len(text) <= max_chars:
        return [text]

    paragraphs = [p.strip() for p in re.split(r"\n+", text) if p.strip()]

    raw_chunks = []
    current = ""

    for paragraph in paragraphs:
        if len(paragraph) > max_chars:
            if current:
                raw_chunks.append(current.strip())
                current = ""

            raw_chunks.extend(split_long_paragraph(paragraph, max_chars=max_chars))
            continue

        if len(current) + len(paragraph) + 2 <= max_chars:
            current = f"{current}\n\n{paragraph}".strip()
        else:
            if current:
                raw_chunks.append(current.strip())
            current = paragraph

    if current:
        raw_chunks.append(current.strip())

    return merge_small_chunks(
    chunks=raw_chunks,
    min_chars=min_chars,
    max_chars=max_chars,
)


def fetch_documents(cursor) -> list[Document]:
    cursor.execute(
        """
        SELECT id, text
        FROM documents
        ORDER BY id;
        """
    )

    rows = cursor.fetchall()

    return [
        Document(id=row[0], text=row[1])
        for row in rows
    ]


def delete_existing_chunks(cursor) -> None:
    cursor.execute("DELETE FROM chunks;")


def insert_chunk(
    cursor,
    document_id: int,
    chunk_number: int,
    content: str,
) -> None:
    content_hash = calculate_hash(content)

    cursor.execute(
        """
        INSERT INTO chunks (
            document_id,
            chunk_number,
            content,
            content_length,
            content_hash
        )
        VALUES (%s, %s, %s, %s, %s)
        ON CONFLICT (document_id, chunk_number)
        DO UPDATE SET
            content = EXCLUDED.content,
            content_length = EXCLUDED.content_length,
            content_hash = EXCLUDED.content_hash,
            opensearch_id = NULL;
        """,
        (
            document_id,
            chunk_number,
            content,
            len(content),
            content_hash,
        ),
    )


def print_stats(cursor) -> None:
    cursor.execute("SELECT COUNT(*) FROM documents;")
    documents_count = cursor.fetchone()[0]

    cursor.execute("SELECT COUNT(*) FROM chunks;")
    chunks_count = cursor.fetchone()[0]

    cursor.execute("SELECT ROUND(AVG(content_length), 2) FROM chunks;")
    avg_chunk_length = cursor.fetchone()[0]

    cursor.execute("SELECT MIN(content_length), MAX(content_length) FROM chunks;")
    min_chunk_length, max_chunk_length = cursor.fetchone()

    cursor.execute(
        """
        SELECT
            d.id,
            COUNT(c.id) AS chunks_count
        FROM documents d
        LEFT JOIN chunks c ON c.document_id = d.id
        GROUP BY d.id
        ORDER BY chunks_count DESC
        LIMIT 5;
        """
    )
    top_documents = cursor.fetchall()

    print("\nChunking stats:")
    print(f"Documents: {documents_count}")
    print(f"Chunks: {chunks_count}")
    print(f"Avg chunk length: {avg_chunk_length}")
    print(f"Min chunk length: {min_chunk_length}")
    print(f"Max chunk length: {max_chunk_length}")

    print("\nTop documents by chunks:")
    for document_id, count in top_documents:
        print(f"- document_id={document_id}: {count} chunks")


def main() -> None:
    parser = argparse.ArgumentParser(description="Chunk documents from PostgreSQL")
    parser.add_argument(
        "--max-chars",
        type=int,
        default=1800,
        help="Maximum characters per chunk",
    )
    parser.add_argument(
        "--min-chars",
        type=int,
        default=300,
        help="Minimum desired chunk size for merging small chunks",
    )
    parser.add_argument(
        "--recreate",
        action="store_true",
        help="Delete all existing chunks before chunking",
    )

    args = parser.parse_args()

    connection = get_connection()

    total_documents = 0
    total_chunks = 0

    try:
        with connection:
            with connection.cursor() as cursor:
                if args.recreate:
                    delete_existing_chunks(cursor)

                documents = fetch_documents(cursor)
                total_documents = len(documents)

                for document in documents:
                    chunks = chunk_text(
                        document.text,
                        max_chars=args.max_chars,
                        min_chars=args.min_chars,
                    )

                    for chunk_number, chunk in enumerate(chunks, start=1):
                        insert_chunk(
                            cursor=cursor,
                            document_id=document.id,
                            chunk_number=chunk_number,
                            content=chunk,
                        )
                        total_chunks += 1

                print_stats(cursor)

    finally:
        connection.close()

    print("\nChunking finished")
    print(f"Processed documents: {total_documents}")
    print(f"Inserted or updated chunks: {total_chunks}")


if __name__ == "__main__":
    main()