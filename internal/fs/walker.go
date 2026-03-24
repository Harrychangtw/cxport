package fs

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type FileEntry struct {
	Path    string // relative path from root
	AbsPath string
	IsDir   bool
}

// Walk collects all files under root, respecting .gitignore patterns.
func Walk(root string) ([]FileEntry, error) {
	ignorePatterns := loadGitignore(root)
	var entries []FileEntry

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors silently
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		// Skip root itself
		if rel == "." {
			return nil
		}

		// Always skip hidden dirs and common noise
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") && base != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip node_modules, vendor, dist, etc.
		if info.IsDir() && isCommonSkipDir(base) {
			return filepath.SkipDir
		}

		// Check gitignore patterns
		if shouldIgnore(rel, info.IsDir(), ignorePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		entries = append(entries, FileEntry{
			Path:    rel,
			AbsPath: path,
			IsDir:   info.IsDir(),
		})

		return nil
	})

	return entries, err
}

// ExpandPaths takes a list of paths (files or dirs) and returns all individual file paths.
func ExpandPaths(root string, paths []string) ([]string, error) {
	var files []string
	seen := make(map[string]bool)

	for _, p := range paths {
		absPath := p
		if !filepath.IsAbs(p) {
			absPath = filepath.Join(root, p)
		}

		info, err := os.Stat(absPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			err := filepath.Walk(absPath, func(path string, fi os.FileInfo, err error) error {
				if err != nil || fi.IsDir() {
					return nil
				}
				base := filepath.Base(path)
				if strings.HasPrefix(base, ".") {
					return nil
				}
				if !seen[path] {
					seen[path] = true
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				continue
			}
		} else {
			if !seen[absPath] {
				seen[absPath] = true
				files = append(files, absPath)
			}
		}
	}

	return files, nil
}

func isCommonSkipDir(name string) bool {
	skip := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
		".git":         true,
		"__pycache__":  true,
		".next":        true,
		".turbo":       true,
		".cache":       true,
	}
	return skip[name]
}

func loadGitignore(root string) []string {
	path := filepath.Join(root, ".gitignore")
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

func shouldIgnore(rel string, isDir bool, patterns []string) bool {
	for _, pattern := range patterns {
		// Simple glob matching
		p := strings.TrimSuffix(pattern, "/")
		matched, _ := filepath.Match(p, filepath.Base(rel))
		if matched {
			return true
		}
		// Also try matching the full relative path
		matched, _ = filepath.Match(p, rel)
		if matched {
			return true
		}
	}
	return false
}

// DetectLanguage returns a language identifier based on file extension.
func DetectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	langs := map[string]string{
		".go":    "go",
		".js":    "javascript",
		".jsx":   "jsx",
		".ts":    "typescript",
		".tsx":   "tsx",
		".py":    "python",
		".rb":    "ruby",
		".rs":    "rust",
		".java":  "java",
		".kt":    "kotlin",
		".swift": "swift",
		".c":     "c",
		".cpp":   "cpp",
		".h":     "c",
		".hpp":   "cpp",
		".cs":    "csharp",
		".php":   "php",
		".html":  "html",
		".css":   "css",
		".scss":  "scss",
		".less":  "less",
		".json":  "json",
		".yaml":  "yaml",
		".yml":   "yaml",
		".toml":  "toml",
		".xml":   "xml",
		".md":    "markdown",
		".sql":   "sql",
		".sh":    "bash",
		".bash":  "bash",
		".zsh":   "zsh",
		".fish":  "fish",
		".vim":   "vim",
		".lua":   "lua",
		".r":     "r",
		".ex":    "elixir",
		".exs":   "elixir",
		".erl":   "erlang",
		".hs":    "haskell",
		".ml":    "ocaml",
		".tf":    "terraform",
		".proto": "protobuf",
		".graphql": "graphql",
		".vue":   "vue",
		".svelte": "svelte",
	}
	if lang, ok := langs[ext]; ok {
		return lang
	}
	return "text"
}
