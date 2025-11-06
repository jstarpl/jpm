package logger

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type LogFileSink struct {
	filePath string
	writer   io.Writer
}

func CreateLogFileSink() (*LogFileSink, error) {
	filePath := getFileName()
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE, os.ModeAppend)
	if err != nil {
		return nil, err
	}

	sink := LogFileSink{
		filePath: filePath,
		writer:   bufio.NewWriter(file),
	}

	return &sink, nil
}

func (s *LogFileSink) Write(data *[]byte) error {
	_, err := s.writer.Write(*data)
	return err
}

func (s *LogFileSink) Close() error {
	return nil
}

func (s *LogFileSink) cycle() {

}

func getFileName() string {
	return fmt.Sprintf("log-%v", "d")
}
