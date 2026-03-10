package files

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/LobovVit/CompareFK/internal/result"
)

type SQLSource struct {
	Name    string
	Path    string
	Content string
}

func ReadFile(fileName string) (string, error) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("os.ReadFile: %w", err)
	}
	return string(file), nil
}

func ResolveSQLSources(dir, glob string, files []string) ([]SQLSource, error) {
	if len(files) > 0 {
		return readSourcesByFiles(files)
	}
	if glob != "" {
		matches, err := filepath.Glob(glob)
		if err != nil {
			return nil, fmt.Errorf("glob %s: %w", glob, err)
		}
		sort.Strings(matches)
		return readSourcesByFiles(matches)
	}
	return readSourcesByDir(dir)
}

func ReadSQLSources(dir, glob string, files []string) ([]string, error) {
	sources, err := ResolveSQLSources(dir, glob, files)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, len(sources))
	for _, src := range sources {
		res = append(res, src.Content)
	}
	return res, nil
}

func readSourcesByFiles(paths []string) ([]SQLSource, error) {
	expanded := make([]string, 0, len(paths))
	for _, p := range paths {
		items, err := expandSQLPath(p)
		if err != nil {
			return nil, err
		}
		expanded = append(expanded, items...)
	}
	unique := uniqueStrings(expanded)
	res := make([]SQLSource, 0, len(unique))
	for _, fileName := range unique {
		content, err := ReadFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", fileName, err)
		}
		res = append(res, SQLSource{
			Name:    filepath.Base(fileName),
			Path:    fileName,
			Content: content,
		})
	}
	return res, nil
}

func expandSQLPath(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}
	if !info.IsDir() {
		return []string{path}, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", path, err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") || !strings.HasSuffix(strings.ToLower(entry.Name()), ".sql") {
			continue
		}
		files = append(files, filepath.Join(path, entry.Name()))
	}
	sort.Slice(files, func(i, j int) bool {
		return naturalLess(filepath.Base(files[i]), filepath.Base(files[j]))
	})
	return files, nil
}

func readSourcesByDir(path string) ([]SQLSource, error) {
	files, err := expandSQLPath(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения каталога %v: %w", path, err)
	}
	return readSourcesByFiles(files)
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	res := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		res = append(res, item)
	}
	return res
}

func naturalLess(a, b string) bool {
	ax := splitNatural(a)
	bx := splitNatural(b)
	for i := 0; i < len(ax) && i < len(bx); i++ {
		if ax[i] == bx[i] {
			continue
		}
		an, aErr := strconv.Atoi(ax[i])
		bn, bErr := strconv.Atoi(bx[i])
		if aErr == nil && bErr == nil {
			return an < bn
		}
		return strings.ToLower(ax[i]) < strings.ToLower(bx[i])
	}
	return len(ax) < len(bx)
}

func splitNatural(s string) []string {
	if s == "" {
		return nil
	}
	parts := make([]string, 0, len(s))
	start := 0
	lastDigit := unicode.IsDigit(rune(s[0]))
	for i, r := range s {
		isDigit := unicode.IsDigit(r)
		if i == 0 {
			continue
		}
		if isDigit != lastDigit {
			parts = append(parts, s[start:i])
			start = i
			lastDigit = isDigit
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func WriteFile(fileName string, lines []string) error {
	path := filepath.Join(result.Res.DateTimeFolder, fileName)
	return WriteFileByPath(path, lines)
}

func WriteFileByPath(path string, lines []string) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return fmt.Errorf("mkdir all: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, 1024*1024)
	for _, line := range lines {
		if _, err := writer.WriteString(line); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
		if _, err := writer.WriteString("\r\n"); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush file: %w", err)
	}
	return nil
}

func WriteSQLFile(fileName string, query string) error {
	path := filepath.Join(result.Res.DateTimeFolder, "sql", fileName)
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return fmt.Errorf("mkdir all: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer file.Close()

	if _, err := file.Write([]byte(query)); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
