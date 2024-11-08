package logger

import (
	"log"
	"os"
	"time"
)

type FileLogger struct {
	file *os.File
	//logger *log.Logger
	buffer *MultiBuffer
	done   chan struct{}
}

func NewFileLogger(filename string, bufferSize int) (*FileLogger, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	f := &FileLogger{
		file: file,
		//logger: log.New(file, "", 0),
		buffer: NewMultiBuffer(bufferSize),
		done:   make(chan struct{}),
	}

	go f.flush()
	return f, nil
}

func (f *FileLogger) Write(message string) {
	f.buffer.Write(message)
}

func (f *FileLogger) flush() {
	ticker := time.NewTicker(time.Second * time.Duration(1))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logs := f.buffer.Read()
			//fmt.Println(logs)
			if logs != "" {
				_, err := f.file.Write([]byte(logs))
				if err != nil {
					log.Printf("Failed to write logs to file: %v", err)
				}

				err = f.file.Sync()
				if err != nil {
					log.Printf("Failed to sync logs to file: %v", err)
				}
			}
		case <-f.done:
			return
		}
	}
}

func (f *FileLogger) Close() {
	close(f.done)
	f.file.Close()
}
