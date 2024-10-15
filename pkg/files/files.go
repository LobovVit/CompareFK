package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LobovVit/CompareFK/internal/result"
)

func ReadFile(fileName string) (string, error) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("os.ReadFile: %w", err)
	}
	ret := string(file)
	return ret, nil
}

func ReadCatalog(path string) ([]string, error) {
	res := make([]string, 0)
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения каталога %v: %w", path, err)
	}

	for _, file := range files {
		script, err := ReadFile(path + file.Name())
		if err != nil {
			continue
		}
		res = append(res, script)
	}
	return res, nil
}

func WriteFile(fileName string, slice []string) error {
	s := strings.Join(slice[0:], "\r\n")
	path := filepath.Join(result.Res.DateTimeFolder, fileName)
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	_, err = file.Write([]byte(s))
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("close file: %w", err)
	}
	return nil
}

func WriteSQLFile(fileName string, query string) error {
	path := filepath.Join(result.Res.DateTimeFolder, "sql", fileName)
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	_, err = file.Write([]byte(query))
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("close file: %w", err)
	}
	return nil
}
