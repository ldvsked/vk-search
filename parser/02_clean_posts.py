import argparse
import hashlib
import json
import re
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


def read_jsonl(path: Path) -> list[dict[str, Any]]:
    rows = []

    if not path.exists():
        raise FileNotFoundError(f"File not found: {path}")

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


def write_jsonl(path: Path, rows: list[dict[str, Any]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)

    with path.open("w", encoding="utf-8") as file:
        for row in rows:
            file.write(json.dumps(row, ensure_ascii=False) + "\n")


def read_report(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {}

    with path.open("r", encoding="utf-8") as file:
        return json.load(file)


def write_report(path: Path, report: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)

    with path.open("w", encoding="utf-8") as file:
        json.dump(report, file, ensure_ascii=False, indent=2)


def normalize_text(text: str) -> str:
    """
    Очищает текст поста без агрессивного удаления смысла.
    Важно: мы не удаляем весь текст после ссылок и не режем предложения,
    потому что это будет использоваться для поиска и LLM.
    """
    if not text:
        return ""

    text = text.replace("\xa0", " ")

    # Удаляем ссылки, но оставляем остальной текст.
    text = re.sub(r"https?://\S+", " ", text)
    text = re.sub(r"www\.\S+", " ", text)

    # Убираем VK-упоминания вида [club123|текст] или [id123|имя],
    # оставляем только человекочитаемую часть после |
    text = re.sub(r"\[(?:club|public|id)\d+\|([^\]]+)\]", r"\1", text)

    # Нормализуем переносы строк.
    text = re.sub(r"\r\n", "\n", text)
    text = re.sub(r"\n{3,}", "\n\n", text)

    # Нормализуем пробелы внутри строк.
    lines = []
    for line in text.split("\n"):
        line = re.sub(r"[ \t]+", " ", line).strip()
        if line:
            lines.append(line)

    return "\n".join(lines).strip()


def calculate_hash(text: str) -> str:
    return hashlib.sha256(text.encode("utf-8")).hexdigest()


def build_title(row: dict[str, Any], clean_text: str) -> str:
    published_at = row.get("published_at")
    source_name = row.get("source_name") or "VK"

    if published_at:
        date_part = published_at[:10]
        return f"Пост {source_name} от {date_part}"

    first_line = clean_text.split("\n")[0].strip()
    if first_line:
        return first_line[:120]

    return f"Пост {source_name}"


def clean_posts(
    raw_posts: list[dict[str, Any]],
    min_length: int,
) -> tuple[list[dict[str, Any]], dict[str, Any]]:
    clean_rows = []

    seen_external_ids = set()
    seen_hashes = set()

    removed_empty = 0
    removed_short = 0
    removed_duplicate_external_id = 0
    removed_duplicate_hash = 0

    for row in raw_posts:
        external_id = row.get("external_id")

        if external_id in seen_external_ids:
            removed_duplicate_external_id += 1
            continue

        seen_external_ids.add(external_id)

        raw_text = row.get("text") or ""
        clean_text = normalize_text(raw_text)

        if not clean_text:
            removed_empty += 1
            continue

        if len(clean_text) < min_length:
            removed_short += 1
            continue

        content_hash = calculate_hash(clean_text)

        if content_hash in seen_hashes:
            removed_duplicate_hash += 1
            continue

        seen_hashes.add(content_hash)

        clean_rows.append(
            {
                "source_name": row.get("source_name"),
                "source_domain": row.get("source_domain"),
                "external_id": external_id,
                "title": build_title(row, clean_text),
                "text": clean_text,
                "content_hash": content_hash,
                "url": row.get("url"),
                "published_at": row.get("published_at"),
                "raw_json": row.get("raw_json", {}),
            }
        )

    lengths = [len(row["text"]) for row in clean_rows]

    stats = {
        "total_raw_posts": len(raw_posts),
        "clean_documents": len(clean_rows),
        "removed_empty": removed_empty,
        "removed_short": removed_short,
        "removed_duplicate_external_id": removed_duplicate_external_id,
        "removed_duplicate_hash": removed_duplicate_hash,
        "min_text_length": min(lengths) if lengths else 0,
        "max_text_length": max(lengths) if lengths else 0,
        "avg_text_length": round(sum(lengths) / len(lengths), 2) if lengths else 0,
    }

    return clean_rows, stats


def main() -> None:
    parser = argparse.ArgumentParser(description="Clean raw VK posts")
    parser.add_argument(
        "--input",
        type=str,
        default="data/raw_posts.jsonl",
        help="Path to raw posts JSONL",
    )
    parser.add_argument(
        "--output",
        type=str,
        default="data/clean_posts.jsonl",
        help="Path to clean posts JSONL",
    )
    parser.add_argument(
        "--report",
        type=str,
        default="data/parse_report.json",
        help="Path to parse report JSON",
    )
    parser.add_argument(
        "--min-length",
        type=int,
        default=100,
        help="Minimum text length after cleaning",
    )

    args = parser.parse_args()

    input_path = Path(args.input)
    output_path = Path(args.output)
    report_path = Path(args.report)

    raw_posts = read_jsonl(input_path)
    clean_rows, clean_stats = clean_posts(raw_posts, min_length=args.min_length)

    write_jsonl(output_path, clean_rows)

    report = read_report(report_path)
    report["cleaning"] = {
        "started_at": datetime.now(timezone.utc).isoformat(),
        "input": args.input,
        "output": args.output,
        "min_length": args.min_length,
        **clean_stats,
    }
    write_report(report_path, report)

    print("Cleaning finished")
    print(f"Raw posts: {clean_stats['total_raw_posts']}")
    print(f"Clean documents: {clean_stats['clean_documents']}")
    print(f"Removed empty: {clean_stats['removed_empty']}")
    print(f"Removed short: {clean_stats['removed_short']}")
    print(f"Removed duplicate external_id: {clean_stats['removed_duplicate_external_id']}")
    print(f"Removed duplicate hash: {clean_stats['removed_duplicate_hash']}")
    print(f"Avg text length: {clean_stats['avg_text_length']}")
    print(f"Output: {args.output}")
    print(f"Report: {args.report}")


if __name__ == "__main__":
    main()