-- Bound wiki page version history and prepare retained versions for compressed storage.
ALTER TABLE wiki_page_versions ADD COLUMN content_encoding TEXT NOT NULL DEFAULT 'plain';
ALTER TABLE wiki_page_versions ADD COLUMN content_compressed BLOB;

CREATE INDEX IF NOT EXISTS idx_wiki_page_versions_page_version_desc
    ON wiki_page_versions(wiki_page_id, version_number DESC);

DELETE FROM wiki_page_versions
WHERE id IN (
    SELECT id
    FROM (
        SELECT
            id,
            ROW_NUMBER() OVER (
                PARTITION BY wiki_page_id
                ORDER BY version_number DESC
            ) AS rn
        FROM wiki_page_versions
    ) ranked
    WHERE rn > 50
);
