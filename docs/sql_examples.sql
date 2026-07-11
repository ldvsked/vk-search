-- 1. Количество документов по источникам
SELECT
    s.name AS source_name,
    s.domain,
    COUNT(d.id) AS documents_count
FROM sources s
LEFT JOIN documents d ON d.source_id = s.id
GROUP BY s.id, s.name, s.domain
ORDER BY documents_count DESC;


-- 2. Количество чанков по источникам
SELECT
    s.name AS source_name,
    COUNT(c.id) AS chunks_count,
    ROUND(AVG(c.content_length), 2) AS avg_chunk_length
FROM sources s
JOIN documents d ON d.source_id = s.id
JOIN chunks c ON c.document_id = d.id
GROUP BY s.id, s.name
ORDER BY chunks_count DESC;


-- 3. Самые длинные документы
SELECT
    d.id,
    s.domain,
    d.title,
    LENGTH(d.text) AS document_length,
    d.url
FROM documents d
JOIN sources s ON s.id = d.source_id
ORDER BY document_length DESC
LIMIT 10;


-- 4. Документы, у которых больше всего чанков
SELECT
    d.id AS document_id,
    s.domain,
    d.title,
    COUNT(c.id) AS chunks_count,
    d.url
FROM documents d
JOIN sources s ON s.id = d.source_id
JOIN chunks c ON c.document_id = d.id
GROUP BY d.id, s.domain, d.title, d.url
ORDER BY chunks_count DESC
LIMIT 10;


-- 5. Источники, у которых документов больше среднего
SELECT
    source_stats.source_name,
    source_stats.documents_count
FROM (
    SELECT
        s.name AS source_name,
        COUNT(d.id) AS documents_count
    FROM sources s
    JOIN documents d ON d.source_id = s.id
    GROUP BY s.id, s.name
) source_stats
WHERE source_stats.documents_count > (
    SELECT AVG(documents_count)
    FROM (
        SELECT COUNT(d.id) AS documents_count
        FROM sources s
        JOIN documents d ON d.source_id = s.id
        GROUP BY s.id
    ) avg_stats
)
ORDER BY source_stats.documents_count DESC;


-- 6. Последние запуски индексации
SELECT
    id,
    status,
    started_at,
    finished_at,
    documents_count,
    chunks_count,
    error_message
FROM index_runs
ORDER BY started_at DESC
LIMIT 10;


-- 7. Поисковые логи по режимам
SELECT
    mode,
    COUNT(*) AS requests_count,
    ROUND(AVG(result_count), 2) AS avg_result_count
FROM search_logs
GROUP BY mode
ORDER BY requests_count DESC;