# Parser

Python-скрипты для сбора, очистки, загрузки и индексации постов VK.

## День 1

На первом этапе реализован тестовый парсер VK-пабликов.

## Установка зависимостей

```bash
cd parser
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt

## Очистка данных

После получения `data/raw_posts.jsonl` нужно выполнить очистку:

```bash
python parser/02_clean_posts.py