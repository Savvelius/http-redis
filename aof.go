package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

type Aof struct {
	quitch  chan struct{}
	file    *os.File
	written int64
	rd      *bufio.Reader
	mu      sync.RWMutex
}

func NewAof(path string, syncFreq time.Duration) (*Aof, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	fInfo, _ := f.Stat()
	fSize := fInfo.Size()

	aof := &Aof{
		file:    f,
		rd:      bufio.NewReader(f),
		quitch:  make(chan struct{}),
		written: fSize,
	}

	go func() {
		for {
			aof.mu.Lock()
			aof.file.Sync()
			aof.mu.Unlock()

			if _, opened := <-aof.quitch; !opened {
				return
			}

			time.Sleep(syncFreq)
		}
	}()

	return aof, nil
}

func (aof *Aof) IsEmpty() bool {
	aof.mu.RLock()
	defer aof.mu.RUnlock()

	return aof.written == 0
}

func (aof *Aof) Drop() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	close(aof.quitch)
	return aof.file.Close()
}

func (aof *Aof) Write(bts []byte) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	n, err := aof.file.Write(bts)
	aof.written += int64(n)
	return err
}

func (aof *Aof) WriteString(str string) error {
	return aof.Write([]byte(str))
}

// asserts that file started with "[]"
func (aof *Aof) WriteCommand(funcName string, args ...any) error {
	cmd := Command{FuncName: funcName, Args: args}
	bytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	aof.mu.Lock()
	defer aof.mu.Unlock()

	// delete closing bracket
	aof.written -= 1
	aof.file.Truncate(aof.written)
	aof.file.Seek(aof.written, 0)

	// add comma if there are previous array elements
	if aof.written != 1 {
		n, err := aof.file.WriteString(",")
		aof.written += int64(n)
		if err != nil {
			return err
		}
	}

	// write json as another element of array
	n, err := aof.file.Write(bytes)
	aof.written += int64(n)
	if err != nil {
		return err
	}

	// close the array
	n, err = aof.file.WriteString("]")
	aof.written += int64(n)
	if err != nil {
		return err
	}

	return nil
}

func (aof *Aof) Terminate() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	if err := os.Remove(aof.file.Name()); err != nil {
		return err
	}

	close(aof.quitch)
	return nil
}

func (aof *Aof) ReadAll() ([]byte, error) {
	aof.mu.RLock()
	defer aof.mu.RUnlock()

	return io.ReadAll(aof.file)
}

func (aof *Aof) ReadCommands() ([]Command, error) {
	aof.mu.RLock()
	defer aof.mu.RUnlock()

	commands := []Command{}
	bytes, err := io.ReadAll(aof.file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &commands)
	if err != nil {
		return nil, err
	}

	return commands, nil
}
