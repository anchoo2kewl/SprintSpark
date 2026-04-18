// Package blogsync pulls posts tagged with a given category from a remote
// anshumanbiswas.com-compatible blog API and materialises them as markdown
// files (YAML frontmatter + body) for the go-blog engine to serve.
//
// The sync is idempotent: files are only rewritten when the rendered bytes
// differ from what is already on disk. LastEditDate is written into the
// frontmatter, so upstream edits naturally trigger a rewrite.
//
// Go-draw embeds (<div class="godraw-embed" data-src="/draw/{id}" ...>) are
// detected in the source content and replaced with inline <img> tags
// pointing at locally-cached SVGs under <PostsDir>/_images/. The SVGs are
// fetched from <SourceURL>/draw/{id}/export.svg on each sync; writes are
// idempotent, so images that haven't changed upstream are a no-op.
//
// Use Run(ctx, ...) to start a scheduler that syncs immediately and then
// every interval thereafter. Use SyncOnce for a single pass.
package blogsync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Config controls a sync.
type Config struct {
	SourceURL string        // e.g. "https://anshumanbiswas.com"
	Tag       string        // filter: posts with this category (case-insensitive)
	PostsDir  string        // output directory
	UserAgent string        // HTTP User-Agent (Cloudflare blocks some defaults)
	Timeout   time.Duration // per-request timeout
}

// Result summarises a sync pass.
type Result struct {
	Fetched       int // posts pulled from remote
	Matched       int // posts with the configured tag
	Written       int // markdown files created or updated
	Unchanged     int // markdown files skipped because content is identical
	Errors        int // per-post failures
	ImagesFetched int // go-draw SVGs downloaded (new or changed)
	ImagesCached  int // go-draw SVGs unchanged (bytes identical on disk)
}

type remotePost struct {
	ID                int        `json:"ID"`
	Username          string     `json:"Username"`
	AuthorDisplayName string     `json:"AuthorDisplayName"`
	Title             string     `json:"Title"`
	Content           string     `json:"Content"`
	Slug              string     `json:"Slug"`
	PublicationDate   string     `json:"PublicationDate"`
	LastEditDate      string     `json:"LastEditDate,omitempty"`
	CreatedAt         time.Time  `json:"CreatedAt"`
	Featured          bool       `json:"Featured"`
	FeaturedImageURL  string     `json:"FeaturedImageURL"`
	Categories        []category `json:"categories"`
}

type category struct {
	Name string `json:"name"`
}

type loadMoreResp struct {
	Posts []remotePost `json:"Posts"`
}

// SyncOnce runs a single sync pass and returns a Result.
func SyncOnce(ctx context.Context, cfg Config) (Result, error) {
	var res Result
	if cfg.SourceURL == "" || cfg.Tag == "" || cfg.PostsDir == "" {
		return res, fmt.Errorf("blogsync: SourceURL, Tag, and PostsDir are required")
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "taskai-blog-sync/1.0"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if err := os.MkdirAll(cfg.PostsDir, 0o755); err != nil {
		return res, fmt.Errorf("create posts dir: %w", err)
	}

	posts, err := fetchAll(ctx, cfg)
	if err != nil {
		return res, err
	}
	res.Fetched = len(posts)

	client := &http.Client{Timeout: cfg.Timeout}
	wantTag := strings.ToLower(strings.TrimSpace(cfg.Tag))
	for _, p := range posts {
		matched := false
		for _, c := range p.Categories {
			if strings.ToLower(c.Name) == wantTag {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		res.Matched++
		// Fetch + persist any go-draw diagrams, then rewrite the content
		// to point at the local copies before rendering the markdown.
		rewritten, imgChanged, imgUnchanged, err := syncDrawings(ctx, client, cfg, p.Content)
		if err != nil {
			res.Errors++
			continue
		}
		res.ImagesFetched += imgChanged
		res.ImagesCached += imgUnchanged
		p.Content = rewritten
		changed, err := writePost(cfg, p)
		if err != nil {
			res.Errors++
			continue
		}
		if changed {
			res.Written++
		} else {
			res.Unchanged++
		}
	}
	return res, nil
}

// godrawEmbedRE matches the exact embed form that anshumanbiswas.com emits
// in post content:
//
//	<div class="godraw-embed" data-src="/draw/<id>" ...></div><script ...></script>
//
// The `data-src` attribute can appear anywhere in the <div> and may use
// single or double quotes. We capture the drawing id for the fetch.
var godrawEmbedRE = regexp.MustCompile(`(?s)<div[^>]*class="godraw-embed"[^>]*data-src=["']/draw/([a-z0-9][a-z0-9\-]*)["'][^>]*>\s*</div>\s*<script[^>]+src=["']/draw/static/embed\.js["'][^>]*>\s*</script>`)

// syncDrawings finds every go-draw embed in content, fetches each drawing
// as an SVG into <PostsDir>/_images/godraw-<id>.svg (transparent background)
// and returns rewritten content where the embeds are replaced with an
// ordinary markdown image reference. The returned counts say how many
// images were actually written vs already up-to-date.
func syncDrawings(ctx context.Context, client *http.Client, cfg Config, content string) (string, int, int, error) {
	matches := godrawEmbedRE.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return content, 0, 0, nil
	}
	imagesDir := filepath.Join(cfg.PostsDir, "_images")
	if err := os.MkdirAll(imagesDir, 0o755); err != nil {
		return content, 0, 0, fmt.Errorf("create images dir: %w", err)
	}
	var (
		out       strings.Builder
		last      int
		fetched   int
		unchanged int
	)
	for _, m := range matches {
		start, end := m[0], m[1]
		idStart, idEnd := m[2], m[3]
		id := content[idStart:idEnd]
		out.WriteString(content[last:start])

		svgBytes, err := fetchDrawingSVG(ctx, client, cfg, id)
		if err != nil {
			// Leave the embed in place -- the post renders without the
			// diagram but the rest of the content still ships.
			out.WriteString(content[start:end])
			last = end
			continue
		}
		imgPath := filepath.Join(imagesDir, "godraw-"+id+".svg")
		changed, err := writeIfChanged(imgPath, svgBytes)
		if err != nil {
			out.WriteString(content[start:end])
			last = end
			continue
		}
		if changed {
			fetched++
		} else {
			unchanged++
		}
		// Markdown image with a stable, absolute path. The server serves
		// /blog-images/* from <PostsDir>/_images/*.
		out.WriteString(fmt.Sprintf(`![Diagram](/blog-images/godraw-%s.svg)`, id))
		last = end
	}
	out.WriteString(content[last:])
	return out.String(), fetched, unchanged, nil
}

// fetchDrawingSVG GETs /draw/{id}/export.svg?bg=transparent from the source
// blog and returns the raw SVG bytes.
func fetchDrawingSVG(ctx context.Context, client *http.Client, cfg Config, id string) ([]byte, error) {
	url := fmt.Sprintf("%s/draw/%s/export.svg?bg=transparent",
		strings.TrimRight(cfg.SourceURL, "/"), id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", cfg.UserAgent)
	req.Header.Set("Accept", "image/svg+xml")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("GET %s: status=%d body=%q", url, resp.StatusCode, string(body))
	}
	return io.ReadAll(io.LimitReader(resp.Body, 4<<20)) // 4 MB hard cap
}

// writeIfChanged writes data to path only if the current on-disk bytes differ.
// Returns (true, nil) when the file was (re-)written, (false, nil) when the
// file is already up to date.
func writeIfChanged(path string, data []byte) (bool, error) {
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, data) {
		return false, nil
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

// Run starts a scheduler that syncs immediately and then every interval.
// It blocks until ctx is cancelled. Errors are logged, never returned --
// a failing sync should not crash the server.
func Run(ctx context.Context, cfg Config, interval time.Duration, logger *zap.Logger) {
	if logger == nil {
		logger = zap.NewNop()
	}
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	runOnce := func() {
		res, err := SyncOnce(ctx, cfg)
		if err != nil {
			logger.Error("blogsync failed", zap.Error(err))
			return
		}
		logger.Info("blogsync completed",
			zap.String("tag", cfg.Tag),
			zap.Int("fetched", res.Fetched),
			zap.Int("matched", res.Matched),
			zap.Int("written", res.Written),
			zap.Int("unchanged", res.Unchanged),
			zap.Int("images_new", res.ImagesFetched),
			zap.Int("images_cached", res.ImagesCached),
			zap.Int("errors", res.Errors),
		)
	}
	// Initial run after a short delay so the main server finishes booting.
	select {
	case <-ctx.Done():
		return
	case <-time.After(5 * time.Second):
	}
	runOnce()
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			runOnce()
		}
	}
}

// fetchAll paginates through /api/posts/load-more until it stops returning
// full pages, with a safety cap to prevent runaway loops.
func fetchAll(ctx context.Context, cfg Config) ([]remotePost, error) {
	client := &http.Client{Timeout: cfg.Timeout}
	const pageSize = 5
	const maxPages = 400
	out := make([]remotePost, 0, 64)
	offset := 0
	for page := 0; page < maxPages; page++ {
		url := fmt.Sprintf("%s/api/posts/load-more?offset=%d", strings.TrimRight(cfg.SourceURL, "/"), offset)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", cfg.UserAgent)
		req.Header.Set("Accept", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("GET %s: %w", url, err)
		}
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			resp.Body.Close()
			return nil, fmt.Errorf("GET %s: status=%d body=%q", url, resp.StatusCode, string(body))
		}
		var lm loadMoreResp
		if err := json.NewDecoder(resp.Body).Decode(&lm); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode: %w", err)
		}
		resp.Body.Close()
		if len(lm.Posts) == 0 {
			break
		}
		out = append(out, lm.Posts...)
		offset += len(lm.Posts)
		if len(lm.Posts) < pageSize {
			break
		}
	}
	return out, nil
}

// writePost serializes a remotePost to markdown + YAML frontmatter. It writes
// the file only if the rendered bytes differ from the current on-disk file,
// so repeated runs with no upstream changes are no-ops.
func writePost(cfg Config, p remotePost) (bool, error) {
	if p.Slug == "" {
		return false, fmt.Errorf("empty slug")
	}
	path := filepath.Join(cfg.PostsDir, p.Slug+".md")
	newData := renderMarkdown(cfg, p)
	if existing, err := os.ReadFile(path); err == nil && bytesEqual(existing, newData) {
		return false, nil
	}
	if err := os.WriteFile(path, newData, 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func renderMarkdown(cfg Config, p remotePost) []byte {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %s\n", yamlString(strings.TrimSpace(p.Title)))
	if author := strings.TrimSpace(p.AuthorDisplayName); author != "" {
		fmt.Fprintf(&b, "author: %s\n", yamlString(author))
	} else if p.Username != "" {
		fmt.Fprintf(&b, "author: %s\n", yamlString(p.Username))
	}
	if t := parsePubDate(p); !t.IsZero() {
		fmt.Fprintf(&b, "date: %s\n", t.Format("2006-01-02"))
	}
	if updated := parseISO(p.LastEditDate); !updated.IsZero() {
		fmt.Fprintf(&b, "updated: %s\n", updated.Format("2006-01-02"))
	}
	// Author URL points to the user's profile on the source blog, so the
	// clickable author name on taskai.cc/blog leads back to the original
	// site for attribution and SEO.
	if p.Username != "" {
		fmt.Fprintf(&b, "author_url: %s\n",
			yamlString(strings.TrimRight(cfg.SourceURL, "/")+"/users/"+p.Username))
	}
	if cover := resolveCover(cfg, p.FeaturedImageURL); cover != "" {
		fmt.Fprintf(&b, "cover: %s\n", yamlString(cover))
	}
	if tags := collectTags(p.Categories); len(tags) > 0 {
		quoted := make([]string, len(tags))
		for i, t := range tags {
			quoted[i] = yamlString(t)
		}
		fmt.Fprintf(&b, "tags: [%s]\n", strings.Join(quoted, ", "))
	}
	if p.Featured {
		b.WriteString("featured: true\n")
	}
	// Canonical + source both point at the original post. Canonical is the
	// signal search engines use to avoid duplicate-content penalties; source
	// is shown as a visible back-link on the page.
	sourcePostURL := strings.TrimRight(cfg.SourceURL, "/") + "/blog/" + p.Slug
	fmt.Fprintf(&b, "canonical: %s\n", yamlString(sourcePostURL))
	fmt.Fprintf(&b, "source: %s\n", yamlString(sourcePostURL))
	fmt.Fprintf(&b, "source_id: %d\n", p.ID)
	b.WriteString("---\n\n")
	b.WriteString(cleanContent(p.Content))
	if !strings.HasSuffix(p.Content, "\n") {
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

// moreMarker matches the <more--> excerpt separator on its own line.
var moreMarker = regexp.MustCompile(`(?m)^[ \t]*<more-->[ \t]*\n?`)

func cleanContent(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = moreMarker.ReplaceAllString(s, "")
	return strings.TrimSpace(s) + "\n"
}

func collectTags(cats []category) []string {
	out := make([]string, 0, len(cats))
	for _, c := range cats {
		name := strings.TrimSpace(c.Name)
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

func resolveCover(cfg Config, s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	return strings.TrimRight(cfg.SourceURL, "/") + s
}

func parsePubDate(p remotePost) time.Time {
	for _, l := range []string{"January 2, 2006", "Jan 2, 2006", "2006-01-02", time.RFC3339} {
		if t, err := time.Parse(l, p.PublicationDate); err == nil {
			return t
		}
	}
	if !p.CreatedAt.IsZero() {
		return p.CreatedAt
	}
	return time.Time{}
}

func parseISO(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	for _, l := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02T15:04:05.000000Z"} {
		if t, err := time.Parse(l, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func yamlString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return "\"" + s + "\""
}
