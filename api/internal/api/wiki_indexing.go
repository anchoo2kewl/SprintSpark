package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"taskai/ent"
	"taskai/ent/pageversion"
	"taskai/ent/wikiblock"
	"taskai/ent/wikipage"
	"taskai/ent/wikipageversion"
)

// StartIndexingWorker starts a background worker that periodically indexes wiki content
func (s *Server) StartIndexingWorker(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	s.logger.Info("Starting wiki indexing worker",
		zap.Duration("interval", 2*time.Minute),
	)

	// Run immediately on startup
	s.indexPages(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Indexing worker shutting down")
			return
		case <-ticker.C:
			s.indexPages(ctx)
		}
	}
}

// indexPages finds pages that need indexing and indexes them
func (s *Server) indexPages(parentCtx context.Context) {
	ctx, cancel := context.WithTimeout(parentCtx, 2*time.Minute)
	defer cancel()

	// Find pages that have been updated recently (within last 3 minutes)
	pages, err := s.db.Client.WikiPage.Query().
		Where(wikipage.UpdatedAtGTE(time.Now().Add(-3 * time.Minute))).
		All(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch pages for indexing",
			zap.Error(err),
		)
		return
	}

	if len(pages) == 0 {
		s.logger.Debug("No pages need indexing")
		return
	}

	s.logger.Info("Indexing pages",
		zap.Int("page_count", len(pages)),
	)

	successCount := 0
	failCount := 0

	for _, page := range pages {
		if err := s.indexPage(ctx, page); err != nil {
			s.logger.Error("Failed to index page",
				zap.Int64("page_id", page.ID),
				zap.String("page_title", page.Title),
				zap.Error(err),
			)
			failCount++
		} else {
			successCount++
		}
	}

	s.logger.Info("Indexing completed",
		zap.Int("success", successCount),
		zap.Int("failed", failCount),
	)
}

// indexPage indexes a single wiki page
func (s *Server) indexPage(ctx context.Context, page *ent.WikiPage) error {
	// Try Yjs-based block extraction first (real-time editing pathway)
	blocks, err := s.extractBlocksFromYjs(ctx, page)
	if err != nil {
		s.logger.Debug("Yjs extraction failed, trying markdown fallback",
			zap.Int64("page_id", page.ID),
			zap.Error(err),
		)
	}

	// Fallback: use markdown content from wiki_page_versions
	if len(blocks) == 0 {
		blocks, err = s.extractBlocksFromMarkdown(ctx, page)
		if err != nil {
			return fmt.Errorf("extract blocks: %w", err)
		}
	}

	// Delete existing blocks for this page
	_, deleteErr := s.db.Client.WikiBlock.Delete().
		Where(wikiblock.PageID(page.ID)).
		Exec(ctx)
	if deleteErr != nil {
		return deleteErr
	}

	// Insert new blocks
	var savedBlocks []*ent.WikiBlock
	if len(blocks) > 0 {
		savedBlocks, err = s.db.Client.WikiBlock.CreateBulk(blocks...).Save(ctx)
		if err != nil {
			return err
		}
	}

	s.logger.Info("Indexed page",
		zap.Int64("page_id", page.ID),
		zap.String("page_title", page.Title),
		zap.Int("block_count", len(savedBlocks)),
	)

	// Generate and store embeddings (skip if embedding client is nil)
	if s.embeddingClient != nil && len(savedBlocks) > 0 {
		if err := s.embedBlocks(ctx, savedBlocks); err != nil {
			s.logger.Warn("Failed to embed blocks (non-fatal)",
				zap.Int64("page_id", page.ID),
				zap.Error(err),
			)
		}
	}

	return nil
}

// extractBlocksFromYjs attempts to extract blocks via the Yjs processor (binary state).
func (s *Server) extractBlocksFromYjs(ctx context.Context, page *ent.WikiPage) ([]*ent.WikiBlockCreate, error) {
	snapshot, err := s.db.Client.PageVersion.Query().
		Where(pageversion.PageID(page.ID)).
		Order(ent.Desc(pageversion.FieldVersionNumber)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	stateBase64 := base64.StdEncoding.EncodeToString(snapshot.YjsState)
	blocks, err := s.yjsClient.ExtractBlocks(ctx, stateBase64)
	if err != nil {
		return nil, err
	}

	bulk := make([]*ent.WikiBlockCreate, len(blocks))
	for i, block := range blocks {
		bulk[i] = s.db.Client.WikiBlock.Create().
			SetPageID(page.ID).
			SetBlockType(block.Type).
			SetHeadingsPath(block.HeadingsPath).
			SetPlainText(block.PlainText).
			SetPosition(block.Position)
		if block.Level != nil {
			bulk[i].SetLevel(*block.Level)
		}
		if block.CanonicalJSON != "" {
			bulk[i].SetCanonicalJSON(block.CanonicalJSON)
		}
	}
	return bulk, nil
}

// extractBlocksFromMarkdown extracts blocks from wiki_page_versions markdown content.
// Splits content by headings for granular search results.
func (s *Server) extractBlocksFromMarkdown(ctx context.Context, page *ent.WikiPage) ([]*ent.WikiBlockCreate, error) {
	version, err := s.db.Client.WikiPageVersion.Query().
		Where(wikipageversion.WikiPageID(page.ID)).
		Order(ent.Desc(wikipageversion.FieldVersionNumber)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	content := version.Content
	if content == "" {
		return nil, nil
	}

	// Split markdown by headings into blocks
	lines := strings.Split(content, "\n")
	var bulk []*ent.WikiBlockCreate
	var currentHeading string
	var currentText strings.Builder
	position := 0

	flushBlock := func() {
		text := strings.TrimSpace(currentText.String())
		if text == "" {
			return
		}
		blockType := "paragraph"
		if currentHeading != "" {
			blockType = "section"
		}
		b := s.db.Client.WikiBlock.Create().
			SetPageID(page.ID).
			SetBlockType(blockType).
			SetHeadingsPath(currentHeading).
			SetPlainText(text).
			SetPosition(position)
		bulk = append(bulk, b)
		position++
		currentText.Reset()
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Flush previous block
			flushBlock()
			// Extract heading text (strip # prefix)
			heading := strings.TrimLeft(trimmed, "# ")
			currentHeading = heading
		} else {
			if trimmed != "" {
				if currentText.Len() > 0 {
					currentText.WriteByte('\n')
				}
				currentText.WriteString(trimmed)
			}
		}
	}
	// Flush last block
	flushBlock()

	return bulk, nil
}

// embedBlocks generates and stores vector embeddings for wiki blocks.
func (s *Server) embedBlocks(ctx context.Context, blocks []*ent.WikiBlock) error {
	// Build embedding inputs: headings_path + plain_text for context-enriched vectors
	// Truncate to ~500 chars to stay within model context length (all-minilm max ~256 tokens)
	const maxChars = 500
	texts := make([]string, len(blocks))
	for i, block := range blocks {
		var parts []string
		if block.HeadingsPath != nil && *block.HeadingsPath != "" {
			parts = append(parts, *block.HeadingsPath)
		}
		if block.PlainText != nil && *block.PlainText != "" {
			parts = append(parts, *block.PlainText)
		}
		text := strings.Join(parts, "\n")
		if len(text) > maxChars {
			text = text[:maxChars]
		}
		texts[i] = text
	}

	start := time.Now()
	vectors, err := s.embeddingClient.EmbedBatch(ctx, texts)
	if err != nil {
		return fmt.Errorf("embed batch: %w", err)
	}

	s.logger.Debug("Generated embeddings",
		zap.Int("count", len(vectors)),
		zap.Duration("latency", time.Since(start)),
	)

	// Store embeddings via raw SQL (Ent doesn't support pgvector natively)
	model := s.embeddingClient.Model()
	now := time.Now()
	for i, block := range blocks {
		if vectors[i] == nil {
			continue
		}
		vectorStr := float32SliceToVectorString(vectors[i])
		_, err := s.db.ExecContext(ctx,
			`UPDATE wiki_blocks SET embedding = $1::vector, embedding_model = $2, embedded_at = $3 WHERE id = $4`,
			vectorStr, model, now, block.ID,
		)
		if err != nil {
			s.logger.Warn("Failed to store embedding for block",
				zap.Int64("block_id", block.ID),
				zap.Error(err),
			)
		}
	}

	return nil
}

// HandleReindexWiki triggers re-indexing and embedding of all wiki pages.
// This is useful after deploying embeddings for the first time or after model upgrades.
func (s *Server) HandleReindexWiki(w http.ResponseWriter, r *http.Request) {
	if s.embeddingClient == nil {
		respondError(w, http.StatusServiceUnavailable, "embeddings not configured", "embeddings_disabled")
		return
	}

	// Run in background so the HTTP request returns immediately
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		pages, err := s.db.Client.WikiPage.Query().All(ctx)
		if err != nil {
			s.logger.Error("Reindex: failed to fetch pages", zap.Error(err))
			return
		}

		s.logger.Info("Reindex: starting full wiki re-index",
			zap.Int("page_count", len(pages)),
		)

		success, fail := 0, 0
		for _, page := range pages {
			if err := s.indexPage(ctx, page); err != nil {
				s.logger.Error("Reindex: failed to index page",
					zap.Int64("page_id", page.ID),
					zap.Error(err),
				)
				fail++
			} else {
				success++
			}
		}

		s.logger.Info("Reindex: completed",
			zap.Int("success", success),
			zap.Int("failed", fail),
		)
	}()

	respondJSON(w, http.StatusAccepted, map[string]string{
		"status":  "accepted",
		"message": "re-indexing started in background",
	})
}

// float32SliceToVectorString converts a float32 slice to pgvector string format: "[0.1,0.2,0.3]"
func float32SliceToVectorString(v []float32) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(fmt.Sprintf("%g", f))
	}
	b.WriteByte(']')
	return b.String()
}
