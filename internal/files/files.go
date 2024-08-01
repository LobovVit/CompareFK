package files

import (
	"fmt"
	"os"
	"strings"

	"Compare/pkg/logger"
	"go.uber.org/zap"
)

func ReadFile(fileName string) (string, error) {
	file, err := os.ReadFile(fileName)
	if err != nil {
		logger.Log.Error("Ошибка чтения файла", zap.String("fileName", fileName), zap.Error(err))
		return "", err
	}
	ret := string(file)
	logger.Log.Info(ret)
	return ret, nil
}

//func WriteFile3(fileName string, slice []string) error {
//	s := strings.Join(slice[0:], "\r\n")
//	err := os.WriteFile(fileName, []byte(s), 0644)
//	if err != nil {
//		logger.Log.Error("ERR", zap.Error(err))
//		return fmt.Errorf("WriteFile: %w", err)
//	}
//	logger.Log.Info("Файл записан", zap.String("файл", fileName))
//	return nil
//}

func WriteFile(fileName string, slice []string) error {
	s := strings.Join(slice[0:], "\r\n")
	file, err := os.Create(fileName)
	if err != nil {
		logger.Log.Error("os.Create", zap.Error(err))
		return fmt.Errorf("os.Create: %w", err)
	}
	_, err = file.Write([]byte(s))
	if err != nil {
		logger.Log.Error("ERR", zap.Error(err))
		return fmt.Errorf("write file: %w", err)
	}
	err = file.Close()
	if err != nil {
		logger.Log.Error("ERR", zap.Error(err))
		return fmt.Errorf("close file: %w", err)
	}
	return nil
}
