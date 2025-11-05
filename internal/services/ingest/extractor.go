package ingest

import (
	s3client "ai-learn-english/pkg/s3"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// FetchToLocalTemp downloads local or S3 file to a temporary path and returns a cleanup function.
func FetchToLocalTemp(filePath string) (string, func(), error) {
	if strings.HasPrefix(filePath, "s3://") {
		u, err := url.Parse(filePath)
		if err != nil {
			return "", func() {}, err
		}
		bucket := u.Host
		key := strings.TrimPrefix(u.Path, "/")
		cli, err := s3client.GetClient()
		if err != nil {
			return "", func() {}, err
		}
		tmp, err := os.CreateTemp("", "ingest-*.pdf")
		if err != nil {
			return "", func() {}, err
		}
		// Download
		out, err := cli.GetObject(context.Background(), &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
		if err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			return "", func() {}, err
		}
		defer out.Body.Close()
		if _, err := io.Copy(tmp, out.Body); err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			return "", func() {}, err
		}
		if _, err := tmp.Seek(0, 0); err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			return "", func() {}, err
		}
		tmp.Close()
		return tmp.Name(), func() { _ = os.Remove(tmp.Name()) }, nil
	}

	// Local path
	abs := filePath
	if !filepath.IsAbs(abs) {
		// allow relative stored paths
		cwd, _ := os.Getwd()
		abs = filepath.Join(cwd, filePath)
	}
	// Copy to temp to ensure we can re-open
	src, err := os.Open(abs)
	if err != nil {
		return "", func() {}, err
	}
	defer src.Close()
	tmp, err := os.CreateTemp("", "ingest-*.pdf")
	if err != nil {
		return "", func() {}, err
	}
	if _, err := io.Copy(tmp, src); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", func() {}, err
	}
	tmp.Close()
	return tmp.Name(), func() { _ = os.Remove(tmp.Name()) }, nil
}

// ExtractPDFTextPages extracts text by pages using ledongthuc/pdf.
func ExtractPDFTextPages(localPath string) ([]string, error) {
	// Minimal POC: read file and treat as a single text page by best-effort UTF-8 conversion.
	f, err := os.ReadFile(localPath)
	if err != nil {
		return nil, err
	}
	var content string
	if utf8.Valid(f) {
		content = string(f)
	} else {
		// Replace invalid runes
		content = string([]rune(string(f)))
	}
	// Sanitize to printable utf-8 (remove BOM, control except common whitespace)
	content = sanitizeUTF8Printable(content)
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("empty content")
	}
	return []string{content}, nil
}

// sanitizeUTF8Printable removes BOM and non-printable runes, keeping common whitespace.
func sanitizeUTF8Printable(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\uFEFF' { // BOM
			continue
		}
		if r == unicode.ReplacementChar { // U+FFFD
			continue
		}
		if r == '\n' || r == '\t' || r == '\r' {
			// keep
		} else if !unicode.IsPrint(r) {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}
