package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhangqiwei/cxport/internal/fs"
)

type Format string

const (
	FormatXML      Format = "xml"
	FormatMarkdown Format = "md"
)

type Result struct {
	OutputPath string
	FileCount  int
	TotalBytes int64
}

// Export concatenates all selected paths into a single output file.
func Export(root string, paths []string, outputPath string, format Format) (Result, error) {
	files, err := fs.ExpandPaths(root, paths)
	if err != nil {
		return Result{}, fmt.Errorf("expanding paths: %w", err)
	}

	var b strings.Builder
	var totalBytes int64

	switch format {
	case FormatMarkdown:
		writeMarkdownHeader(&b)
	default:
		writeXMLHeader(&b)
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		rel, err := filepath.Rel(root, file)
		if err != nil {
			rel = file
		}

		totalBytes += int64(len(content))
		lang := fs.DetectLanguage(file)

		switch format {
		case FormatMarkdown:
			writeMarkdownFile(&b, rel, lang, string(content))
		default:
			writeXMLFile(&b, rel, lang, string(content))
		}
	}

	switch format {
	case FormatMarkdown:
		// no footer needed
	default:
		writeXMLFooter(&b)
	}

	// Determine extension
	ext := ".xml"
	if format == FormatMarkdown {
		ext = ".md"
	}

	finalPath := outputPath + ext
	if err := os.WriteFile(finalPath, []byte(b.String()), 0644); err != nil {
		return Result{}, fmt.Errorf("writing output: %w", err)
	}

	return Result{
		OutputPath: finalPath,
		FileCount:  len(files),
		TotalBytes: totalBytes,
	}, nil
}

func writeXMLHeader(b *strings.Builder) {
	b.WriteString(fmt.Sprintf("<!-- cxport context export | %s -->\n", time.Now().Format("2006-01-02 15:04:05")))
	b.WriteString("<context>\n")
}

func writeXMLFile(b *strings.Builder, path, language, content string) {
	b.WriteString(fmt.Sprintf("<file path=%q language=%q>\n", path, language))
	b.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("</file>\n\n")
}

func writeXMLFooter(b *strings.Builder) {
	b.WriteString("</context>\n")
}

func writeMarkdownHeader(b *strings.Builder) {
	b.WriteString(fmt.Sprintf("# Context Export\n\nGenerated: %s\n\n---\n\n", time.Now().Format("2006-01-02 15:04:05")))
}

func writeMarkdownFile(b *strings.Builder, path, language, content string) {
	b.WriteString(fmt.Sprintf("## File: %s\n\n", path))
	b.WriteString(fmt.Sprintf("```%s\n", language))
	b.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("```\n\n---\n\n")
}
