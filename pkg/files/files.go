package files

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LobovVit/CompareFK/internal/result"
)

func ReadFile(fileName string) (string, error) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("os.ReadFile: %w", err)
	}
	return string(file), nil
}

func ReadSQLSources(dir, glob string, files []string) ([]string, error) {
	if len(files) > 0 {
		return ReadFiles(files)
	}
	if glob != "" {
		matches, err := filepath.Glob(glob)
		if err != nil {
			return nil, fmt.Errorf("glob %s: %w", glob, err)
		}
		sort.Strings(matches)
		return ReadFiles(matches)
	}
	return ReadCatalog(dir)
}

func ReadFiles(fileNames []string) ([]string, error) {
	res := make([]string, 0, len(fileNames))
	for _, fileName := range fileNames {
		content, err := ReadFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", fileName, err)
		}
		res = append(res, content)
	}
	return res, nil
}

func ReadCatalog(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения каталога %v: %w", path, err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	res := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") || !strings.HasSuffix(strings.ToLower(entry.Name()), ".sql") {
			continue
		}
		script, err := ReadFile(filepath.Join(path, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		res = append(res, script)
	}
	return res, nil
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
