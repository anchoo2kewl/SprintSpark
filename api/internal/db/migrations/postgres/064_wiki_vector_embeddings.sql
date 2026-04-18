-- Enable pgvector extension for vector similarity search
CREATE EXTENSION IF NOT EXISTS vector;

-- Add embedding columns to wiki_blocks
ALTER TABLE wiki_blocks ADD COLUMN IF NOT EXISTS embedding vector(384);
ALTER TABLE wiki_blocks ADD COLUMN IF NOT EXISTS embedding_model VARCHAR(100);
ALTER TABLE wiki_blocks ADD COLUMN IF NOT EXISTS embedded_at TIMESTAMPTZ;

-- HNSW index for fast approximate nearest neighbor search (cosine distance)
CREATE INDEX IF NOT EXISTS wiki_blocks_embedding_hnsw
    ON wiki_blocks USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);
