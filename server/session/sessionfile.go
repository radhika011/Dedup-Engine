package session

import (
	"os"
	"sync"
)

type SessionFile struct {
	File *os.File
	Mu   sync.RWMutex
}

func (sessFile *SessionFile) Append(bytes []byte) error {
	sessFile.Mu.Lock()
	_, err := sessFile.File.Write(bytes)
	sessFile.Mu.Unlock()

	return err
}

func (sessFile *SessionFile) Close() error {
	sessFile.Mu.Lock()
	err := sessFile.File.Close()
	sessFile.Mu.Unlock()

	return err
}
