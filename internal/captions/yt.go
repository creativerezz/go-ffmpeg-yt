package captions

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Result struct {
	Source  string // "yt-sub" or "yt-auto-sub"
	Format  string // "text" or "vtt"
	Content string // caption content
}

// FetchWithYtDlp tries to download subtitles using yt-dlp.
// It first tries human-created subtitles, then falls back to auto-generated.
// format can be "text" (default) or "vtt".
func FetchWithYtDlp(ctx context.Context, videoURL, lang, format string) (Result, error) {
	if format == "" {
		format = "text"
	}

	// Ensure yt-dlp exists
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return Result{}, fmt.Errorf("yt-dlp not found in PATH; install it (e.g., pip install yt-dlp): %w", err)
	}

	tempDir, err := os.MkdirTemp("", "yt-captions-")
	if err != nil {
		return Result{}, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Use a deterministic output pattern so we know the filename
	outputPattern := "subs.%(lang)s.%(ext)s"

	// Try authored subtitles first
	vttPath, err := downloadSubs(ctx, tempDir, videoURL, lang, outputPattern, false)
	source := "yt-sub"
	if err != nil || vttPath == "" {
		// Fallback to auto subs
		vttPath, err = downloadSubs(ctx, tempDir, videoURL, lang, outputPattern, true)
		source = "yt-auto-sub"
	}
	if err != nil {
		return Result{}, err
	}
	if vttPath == "" {
		return Result{}, errors.New("no subtitles found for requested language")
	}

	data, err := os.ReadFile(vttPath)
	if err != nil {
		return Result{}, fmt.Errorf("read vtt: %w", err)
	}

	res := Result{Source: source, Format: format}
	if strings.ToLower(format) == "vtt" {
		res.Content = string(data)
		return res, nil
	}

	res.Content = vttToText(string(data))
	return res, nil
}

func downloadSubs(ctx context.Context, dir, url, lang, outPattern string, auto bool) (string, error) {
	args := []string{"--no-progress", "--skip-download", "--sub-format", "vtt", "--sub-langs", lang, "-o", outPattern}
	if auto {
		args = append(args, "--write-auto-sub")
	} else {
		args = append(args, "--write-sub")
	}
	args = append(args, url)

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Add a timeout guard if the parent context doesn't have one
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) <= 0 {
		c2, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		cmd = exec.CommandContext(c2, "yt-dlp", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w", err)
	}

	// Expect file: subs.<lang>.vtt
	candidate := filepath.Join(dir, fmt.Sprintf("subs.%s.vtt", lang))
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	// If language variants exist (e.g., en-US), pick the first matching subs.*.vtt
	var found string
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasPrefix(filepath.Base(path), "subs.") && strings.HasSuffix(path, ".vtt") {
			found = path
			return fs.SkipDir
		}
		return nil
	})
	return found, nil
}

// vttToText provides a minimal conversion of VTT to plain text by removing timestamps,
// cue numbers, and headers. It does not preserve exact timings or speaker labels.
func vttToText(vtt string) string {
	lines := strings.Split(vtt, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		if strings.HasPrefix(l, "WEBVTT") || strings.HasPrefix(l, "NOTE") || strings.HasPrefix(l, "STYLE") {
			continue
		}
		// Timestamp lines like 00:00:01.000 --> 00:00:03.000
		if strings.Contains(l, " --> ") {
			continue
		}
		// Cue numbers are numeric-only lines
		if isAllDigits(l) {
			continue
		}
		out = append(out, l)
	}
	// Merge and de-duplicate adjacent duplicates caused by overlapping cues
	merged := make([]string, 0, len(out))
	var prev string
	for _, l := range out {
		if l == prev {
			continue
		}
		merged = append(merged, l)
		prev = l
	}
	return strings.Join(merged, "\n")
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

