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
```

## Очистка данных

После получения `data/raw_posts.jsonl` нужно выполнить очистку:

```bash
python parser/02_clean_posts.py
```


## Чанкинг документов

После загрузки очищенных документов в PostgreSQL выполняется разбиение текстов на чанки:

```bash
python parser/04_chunk_documents.py --recreate
```


## Создание индекса OpenSearch

Перед индексацией нужно создать индекс:

```bash
python parser/05_create_opensearch_index.py --recreate
```