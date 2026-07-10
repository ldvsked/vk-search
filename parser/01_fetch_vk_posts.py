import argparse
import json
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

import requests
from tqdm import tqdm

from config import get_settings


VK_WALL_GET_URL = "https://api.vk.com/method/wall.get"


def unix_to_iso(timestamp: int | None) -> str | None:
    if not timestamp:
        return None
    return datetime.fromtimestamp(timestamp, tz=timezone.utc).isoformat()


def normalize_domain(public: str) -> str:
    public = public.strip()
    if public.startswith("https://vk.com/"):
        public = public.replace("https://vk.com/", "")
    if public.startswith("vk.com/"):
        public = public.replace("vk.com/", "")
    if public.startswith("@"):
        public = public[1:]
    return public.strip("/")


def build_post_url(owner_id: int, post_id: int) -> str:
    return f"https://vk.com/wall{owner_id}_{post_id}"


def fetch_posts_for_public(
    domain: str,
    token: str,
    api_version: str,
    target_posts: int,
    sleep_seconds: float,
) -> list[dict[str, Any]]:
    posts: list[dict[str, Any]] = []
    offset = 0
    count = 100

    progress = tqdm(total=target_posts, desc=f"Fetch {domain}")

    while len(posts) < target_posts:
        params = {
            "domain": domain,
            "count": count,
            "offset": offset,
            "access_token": token,
            "v": api_version,
        }

        response = requests.get(VK_WALL_GET_URL, params=params, timeout=30)
        response.raise_for_status()

        payload = response.json()

        if "error" in payload:
            error = payload["error"]
            raise RuntimeError(
                f"VK API error for {domain}: "
                f"{error.get('error_code')} {error.get('error_msg')}"
            )

        items = payload.get("response", {}).get("items", [])

        if not items:
            break

        for item in items:
            owner_id = item.get("owner_id")
            post_id = item.get("id")
            text = item.get("text") or ""

            if owner_id is None or post_id is None:
                continue

            posts.append(
                {
                    "source_name": domain,
                    "source_domain": domain,
                    "external_id": f"{owner_id}_{post_id}",
                    "owner_id": owner_id,
                    "post_id": post_id,
                    "text": text,
                    "url": build_post_url(owner_id, post_id),
                    "published_at": unix_to_iso(item.get("date")),
                    "raw_json": item,
                }
            )

            progress.update(1)

            if len(posts) >= target_posts:
                break

        offset += count
        time.sleep(sleep_seconds)

    progress.close()
    return posts


def save_jsonl(path: Path, rows: list[dict[str, Any]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)

    with path.open("w", encoding="utf-8") as file:
        for row in rows:
            file.write(json.dumps(row, ensure_ascii=False) + "\n")


def save_report(path: Path, report: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)

    with path.open("w", encoding="utf-8") as file:
        json.dump(report, file, ensure_ascii=False, indent=2)


def main() -> None:
    parser = argparse.ArgumentParser(description="Fetch posts from VK publics")
    parser.add_argument(
        "--limit-per-public",
        type=int,
        default=None,
        help="How many posts to fetch from each public",
    )
    parser.add_argument(
        "--output",
        type=str,
        default="data/raw_posts.jsonl",
        help="Path to output JSONL file",
    )
    parser.add_argument(
        "--report",
        type=str,
        default="data/parse_report.json",
        help="Path to report JSON file",
    )
    args = parser.parse_args()

    settings = get_settings()

    if not settings.vk_token:
        raise ValueError("VK_TOKEN is empty. Put your VK token into .env")

    if not settings.vk_publics:
        raise ValueError("VK_PUBLICS is empty. Add publics into .env")

    limit_per_public = args.limit_per_public or settings.vk_target_posts

    all_posts: list[dict[str, Any]] = []
    errors: list[dict[str, str]] = []

    started_at = datetime.now(timezone.utc).isoformat()

    for public in settings.vk_publics:
        domain = normalize_domain(public)

        try:
            posts = fetch_posts_for_public(
                domain=domain,
                token=settings.vk_token,
                api_version=settings.vk_api_version,
                target_posts=limit_per_public,
                sleep_seconds=settings.vk_request_sleep_seconds,
            )
            all_posts.extend(posts)
        except Exception as exc:
            errors.append({"public": domain, "error": str(exc)})

    save_jsonl(Path(args.output), all_posts)

    report = {
        "started_at": started_at,
        "finished_at": datetime.now(timezone.utc).isoformat(),
        "target_per_public": limit_per_public,
        "publics": settings.vk_publics,
        "total_raw_posts": len(all_posts),
        "errors": errors,
        "output": args.output,
    }

    save_report(Path(args.report), report)

    print(f"Saved raw posts: {len(all_posts)}")
    print(f"Output: {args.output}")
    print(f"Report: {args.report}")

    if errors:
        print("Errors:")
        for error in errors:
            print(f"- {error['public']}: {error['error']}")


if __name__ == "__main__":
    main()