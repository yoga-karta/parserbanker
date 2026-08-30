package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type HistoryEntry struct {
	JobID     string         `json:"job_id"`
	Username  string         `json:"username"`
	Timestamp time.Time      `json:"timestamp"`
	Status    string         `json:"status"`
	Error     string         `json:"error,omitempty"`
	Counts    map[string]int `json:"counts,omitempty"`
}

var historyMu sync.Mutex

const historyFile = "/app/data/history.jsonl"

func AppendHistory(entry HistoryEntry) error {
	historyMu.Lock()
	defer historyMu.Unlock()

	_ = os.MkdirAll(filepath.Dir(historyFile), 0755)
	f, err := os.OpenFile(historyFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

func ReadHistory(limit int) ([]HistoryEntry, error) {
	historyMu.Lock()
	defer historyMu.Unlock()

	f, err := os.Open(historyFile)
	if os.IsNotExist(err) {
		return []HistoryEntry{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []HistoryEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry HistoryEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			entries = append(entries, entry)
		}
	}

	if len(entries) > limit {
		return entries[len(entries)-limit:], nil
	}
	return entries, nil
}

func ClearHistory() error {
	historyMu.Lock()
	defer historyMu.Unlock()
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(historyFile)
}
