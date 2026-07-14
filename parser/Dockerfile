FROM python:3.12-slim

WORKDIR /app

COPY parser/requirements.txt /app/parser/requirements.txt

RUN pip install --no-cache-dir -r /app/parser/requirements.txt

COPY parser /app/parser
COPY scripts /app/scripts
COPY data/seed /app/data/seed

ENV PYTHONPATH=/app/parser

WORKDIR /app