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
	// Get the latest snapshot for the page
	snapshot, err := s.db.Client.PageVersion.Query().
		Where(pageversion.PageID(page.ID)).
		Order(ent.Desc(pageversion.FieldVersionNumber)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			s.logger.Debug("No snapshot available for indexing",
				zap.Int64("page_id", page.ID),
			)
			return nil
		}
		return err
	}

	// Convert state to base64 for Yjs processor
	stateBase64 := base64.StdEncoding.EncodeToString(snapshot.YjsState)

	// Extract blocks from the snapshot
	blocks, err := s.yjsClient.ExtractBlocks(ctx, stateBase64)
	if err != nil {
		return err
	}

	// Delete existing blocks for this page
	_, err = s.db.Client.WikiBlock.Delete().
		Where(wikiblock.PageID(page.ID)).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Insert new blocks
	var savedBlocks []*ent.WikiBlock
	if len(blocks) > 0 {
		bulk := make([]*ent.WikiBlockCreate, len(blocks))
		for i, block := range blocks {
			bulk[i] = s.db.Client.WikiBlock.Create().
				SetPageID(page.ID).
				SetBlockType(block.Type).
				SetHeadingsPath(block.HeadingsPath).
				SetPlainText(block.PlainText).
				SetPosition(block.Position)

			// Set optional level for headings
			if block.Level != nil {
				bulk[i].SetLevel(*block.Level)
			}

			// Store canonical JSON as string
			if block.CanonicalJSON != "" {
				bulk[i].SetCanonicalJSON(block.CanonicalJSON)
			}
		}

		savedBlocks, err = s.db.Client.WikiBlock.CreateBulk(bulk...).Save(ctx)
		if err != nil {
			return err
		}
	}

	s.logger.Info("Indexed page",
		zap.Int64("page_id", page.ID),
		zap.String("page_title", page.Title),
		zap.Int("block_count", len(blocks)),
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

// embedBlocks generates and stores vector embeddings for wiki blocks.
func (s *Server) embedBlocks(ctx context.Context, blocks []*ent.WikiBlock) error {
	// Build embedding inputs: headings_path + plain_text for context-enriched vectors
	texts := make([]string, len(blocks))
	for i, block := range blocks {
		var parts []string
		if block.HeadingsPath != nil && *block.HeadingsPath != "" {
			parts = append(parts, *block.HeadingsPath)
		}
		if block.PlainText != nil && *block.PlainText != "" {
			parts = append(parts, *block.PlainText)
		}
		texts[i] = strings.Join(parts, "\n")
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
