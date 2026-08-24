package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type FileObserver struct {
	path    string
	onError func(error)
	mu      sync.Mutex
}

func NewFileObserver(path string, onError func(error)) *FileObserver {
	return &FileObserver{path: path, onError: onError}
}

func (o *FileObserver) ID() string {
	return "file:" + o.path
}

func (o *FileObserver) Update(event Event) {
	if err := o.write(event); err != nil && o.onError != nil {
		o.onError(err)
	}
}

func (o *FileObserver) write(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	file, err := os.OpenFile(o.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("open audit file: %w", err)
	}

	defer file.Close()

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}

	return nil
}
