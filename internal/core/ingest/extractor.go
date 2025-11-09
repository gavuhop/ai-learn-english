package ingest

import (
	s3client "ai-learn-english/pkg/s3"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	pdf "github.com/ledongthuc/pdf"
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

// ExtractPDFTextPages extracts readable text per page using ledongthuc/pdf, avoiding binary PDF streams.
func ExtractPDFTextPages(localPath string) ([]string, error) {
	// Open with PDF parser
	file, reader, err := pdf.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	pageCount := reader.NumPage()
	pages := make([]string, 0, pageCount)
	for pageIndex := 1; pageIndex <= pageCount; pageIndex++ {
		p := reader.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		// Get plain text stream for this page
		plainText, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		text := sanitizeUTF8Printable(plainText)
		if strings.TrimSpace(text) != "" {
			pages = append(pages, text)
		}
	}

	// Fallback to whole-document text if page-wise extraction produced nothing
	if len(pages) == 0 {
		plainDoc, err := reader.GetPlainText()
		if err == nil {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, plainDoc); err == nil {
				text := sanitizeUTF8Printable(buf.String())
				if strings.TrimSpace(text) != "" {
					pages = []string{text}
				}
			}
		}
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("no extractable text")
	}
	return pages, nil
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
