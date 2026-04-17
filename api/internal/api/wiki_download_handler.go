package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"taskai/ent"
	"taskai/ent/wikipage"
)

// drawEmbedRe matches godraw-embed divs in rendered HTML and captures the draw ID.
var drawEmbedRe = regexp.MustCompile(`<div class="godraw-embed"[^>]*data-src="[^"]*?/([a-zA-Z0-9_-]+?)(?:/edit)?"[^>]*></div>`)

// drawScriptRe matches the go-draw embed.js script tags injected by go-wiki.
var drawScriptRe = regexp.MustCompile(`<script[^>]*src="[^"]*embed\.js"[^>]*></script>`)

// pdfCSS contains the inlined styles that replicate the Tailwind Typography
// "prose" class output for light-mode print rendering. Using inlined CSS
// avoids an external CDN dependency and keeps PDF generation deterministic.
const pdfCSS = `
/* ── Base typography (matches Tailwind Typography prose) ── */
body {
  font-family: Inter, -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif;
  font-size: 16px;
  line-height: 1.75;
  color: #374151;
  -webkit-font-smoothing: antialiased;
}

.prose { max-width: none; }

.prose h1 { font-size: 2.25em; margin-top: 0; margin-bottom: 0.8888889em; line-height: 1.1111111; font-weight: 800; color: #111827; }
.prose h2 { font-size: 1.5em; margin-top: 2em; margin-bottom: 1em; line-height: 1.3333333; font-weight: 700; color: #111827; }
.prose h3 { font-size: 1.25em; margin-top: 1.6em; margin-bottom: 0.6em; line-height: 1.6; font-weight: 600; color: #111827; }
.prose h4 { margin-top: 1.5em; margin-bottom: 0.5em; line-height: 1.5; font-weight: 600; color: #111827; }

.prose p { margin-top: 1.25em; margin-bottom: 1.25em; }
.prose a { color: #5e6ad2; text-decoration: underline; font-weight: 500; }
.prose strong { color: #111827; font-weight: 600; }
.prose em { font-style: italic; }

/* Lists */
.prose ul, ul.list-disc { list-style-type: disc; padding-left: 1.5rem; }
.prose ol, ol.list-decimal { list-style-type: decimal; padding-left: 1.5rem; }
.prose li { margin-top: 0.5em; margin-bottom: 0.5em; }
.space-y-1 > * + * { margin-top: 0.25rem; }

/* Blockquotes */
.prose blockquote,
blockquote.border-l-4 {
  border-left: 4px solid rgba(59, 130, 246, 0.4);
  padding-left: 1rem;
  font-style: italic;
  color: #6b7280;
  margin-top: 1.6em;
  margin-bottom: 1.6em;
}

/* Code */
.prose code { color: #111827; font-weight: 600; font-size: 0.875em; font-family: 'Source Code Pro', Menlo, Monaco, Consolas, monospace; }
.prose code::before { content: none; }
.prose code::after { content: none; }
.prose pre { background-color: #1f2937; color: #e5e7eb; border-radius: 0.375rem; padding: 0.875rem 1.125rem; overflow-x: auto; font-size: 0.875em; line-height: 1.7142857; margin-top: 1.7142857em; margin-bottom: 1.7142857em; }
.prose pre code { background: transparent; color: inherit; font-weight: 400; font-size: inherit; padding: 0; border-radius: 0; }

/* Horizontal rule */
.prose hr { border-color: #e5e7eb; border-top-width: 1px; margin-top: 3em; margin-bottom: 3em; }

/* Tables */
.prose table { border-collapse: collapse; width: 100%; font-size: 0.875em; line-height: 1.7142857; }
.prose thead { border-bottom: 2px solid #d1d5db; }
.prose thead th { padding: 0.625rem 0.75rem; font-weight: 600; text-align: left; color: #111827; }
.prose tbody td { padding: 0.5rem 0.75rem; }
.prose tbody tr { border-bottom: 1px solid #e5e7eb; }

/* Images */
.prose img { margin-top: 2em; margin-bottom: 2em; border-radius: 0.375rem; max-width: 100%; }

/* References (footnotes) */
.prose sup { font-size: 0.75em; }

/* Graph link pills */
a[data-graph-type] {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 1px 8px;
  border-radius: 9999px;
  font-size: 11px;
  font-weight: 500;
  text-decoration: none;
  vertical-align: middle;
}

/* Print-specific */
@media print {
  body { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  .prose pre { break-inside: avoid; }
  .prose table { break-inside: auto; }
  .prose tr { break-inside: avoid; }
  .prose h2, .prose h3, .prose h4 { break-after: avoid; }
  .prose img { break-inside: avoid; }
}

/* Draw diagrams rendered as inline SVG */
.draw-svg-inline { margin: 1.5em 0; max-width: 100%; }
.draw-svg-inline svg { max-width: 100%; height: auto; }

/* Hide unsupported embeds */
.godraw-embed { display: none; }
.figma-embed { display: none; }
`

// pdfHTMLTemplate is the full HTML document template for PDF rendering.
const pdfHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>%s</title>
<style>%s</style>
</head>
<body>
<h1 style="font-size:2em;font-weight:800;color:#111827;margin-bottom:0.5em;padding-bottom:0.5em;border-bottom:2px solid #e5e7eb;">%s</h1>
<div class="prose">%s</div>
</body>
</html>`

// HandleWikiPagePDF generates a PDF of a wiki page that visually matches
// the web rendering. Uses headless Chromium via chromedp.
func (s *Server) HandleWikiPagePDF(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	pageID, err := strconv.ParseInt(chi.URLParam(r, "pageId"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid page ID", "invalid_input")
		return
	}

	page, err := s.fetchWikiPageWithAccess(ctx, userID, pageID)
	if err != nil {
		s.handleWikiPageError(w, err, pageID)
		return
	}

	// Render markdown → HTML using the same pipeline as the preview endpoint.
	renderedHTML := stripDrawEditMode(
		wiki.RenderContent(
			preprocessFigmaShortcodes(
				preprocessGraphLinksForPreview(page.Content),
			),
		),
	)

	// Replace godraw-embed divs with inline transparent SVGs fetched from go-draw.
	renderedHTML = s.inlineDrawSVGs(ctx, renderedHTML)

	htmlDoc := fmt.Sprintf(pdfHTMLTemplate, page.Title, pdfCSS, page.Title, renderedHTML)

	pdfBytes, err := renderHTMLToPDF(ctx, htmlDoc)
	if err != nil {
		s.logger.Error("Failed to generate wiki PDF",
			zap.Int64("page_id", pageID),
			zap.Error(err),
		)
		respondError(w, http.StatusInternalServerError, "failed to generate PDF", "internal_error")
		return
	}

	filename := page.Slug + ".pdf"
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes) //nolint:errcheck
}

// HandleWikiPageMarkdown serves the raw markdown content of a wiki page as
// a downloadable .md file.
func (s *Server) HandleWikiPageMarkdown(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	pageID, err := strconv.ParseInt(chi.URLParam(r, "pageId"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid page ID", "invalid_input")
		return
	}

	pg, err := s.fetchWikiPageWithAccess(ctx, userID, pageID)
	if err != nil {
		s.handleWikiPageError(w, err, pageID)
		return
	}

	filename := pg.Slug + ".md"
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(pg.Content)))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pg.Content)) //nolint:errcheck
}

// ── helpers ─────────────────────────────────────────────────────────

// fetchWikiPageWithAccess loads a wiki page and verifies user access.
func (s *Server) fetchWikiPageWithAccess(ctx context.Context, userID, pageID int64) (*ent.WikiPage, error) {
	pg, err := s.db.Client.WikiPage.Query().
		Where(wikipage.ID(pageID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	hasAccess, err := s.checkProjectAccess(ctx, userID, pg.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("access check: %w", err)
	}
	if !hasAccess {
		return nil, errForbidden
	}
	return pg, nil
}

// handleWikiPageError maps common errors to HTTP responses.
func (s *Server) handleWikiPageError(w http.ResponseWriter, err error, pageID int64) {
	if ent.IsNotFound(err) {
		respondError(w, http.StatusNotFound, "wiki page not found", "not_found")
		return
	}
	if err == errForbidden {
		respondError(w, http.StatusForbidden, "access denied", "forbidden")
		return
	}
	s.logger.Error("Wiki page error", zap.Int64("page_id", pageID), zap.Error(err))
	respondError(w, http.StatusInternalServerError, "internal error", "internal_error")
}

// errForbidden is a sentinel for access-denied conditions.
var errForbidden = fmt.Errorf("forbidden")

// inlineDrawSVGs replaces godraw-embed divs with inline SVG content fetched
// from the go-draw export endpoint. Uses transparent background for clean
// embedding in the PDF.
func (s *Server) inlineDrawSVGs(ctx context.Context, html string) string {
	// Remove embed.js script tags (not needed for static SVGs).
	html = drawScriptRe.ReplaceAllString(html, "")

	// Replace each godraw-embed div with its SVG.
	html = drawEmbedRe.ReplaceAllStringFunc(html, func(match string) string {
		m := drawEmbedRe.FindStringSubmatch(match)
		if len(m) < 2 {
			return match
		}
		drawID := m[1]

		svgContent, err := s.fetchDrawSVG(ctx, drawID)
		if err != nil {
			s.logger.Warn("Failed to fetch draw SVG for PDF",
				zap.String("draw_id", drawID),
				zap.Error(err),
			)
			return "" // silently omit broken diagrams
		}

		return fmt.Sprintf(`<div class="draw-svg-inline">%s</div>`, svgContent)
	})

	return html
}

// fetchDrawSVG fetches the SVG export of a drawing from the local go-draw
// endpoint with a transparent background.
func (s *Server) fetchDrawSVG(ctx context.Context, drawID string) (string, error) {
	url := fmt.Sprintf("http://localhost:%s/draw/%s/export.svg?bg=transparent", s.config.Port, drawID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch SVG: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("SVG export returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read SVG body: %w", err)
	}

	// Strip the XML declaration if present — we're embedding inline.
	svg := strings.TrimSpace(string(body))
	if strings.HasPrefix(svg, "<?xml") {
		if idx := strings.Index(svg, "?>"); idx != -1 {
			svg = strings.TrimSpace(svg[idx+2:])
		}
	}

	return svg, nil
}

// renderHTMLToPDF launches headless Chromium to print the given HTML to PDF.
func renderHTMLToPDF(ctx context.Context, htmlContent string) ([]byte, error) {
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer allocCancel()

	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	defer taskCancel()

	var pdfBuf []byte
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return fmt.Errorf("get frame tree: %w", err)
			}
			return page.SetDocumentContent(frameTree.Frame.ID, htmlContent).Do(ctx)
		}),
		// Give the browser a moment to lay out styles.
		chromedp.Sleep(200*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithMarginTop(0.6).
				WithMarginBottom(0.6).
				WithMarginLeft(0.5).
				WithMarginRight(0.5).
				WithPreferCSSPageSize(true).
				Do(ctx)
			return err
		}),
	); err != nil {
		return nil, fmt.Errorf("chromedp run: %w", err)
	}

	return pdfBuf, nil
}
