package indexer

import (
	"crypto/sha256"
	"duplication-finder/logconfig"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileInfo struct {
	Path string
	Hash string
}

type Indexer interface {
	fileHash(filePath string) (string, error)
	buildFileIndex(rootPath string) error
	GetFileIndex(rootDir string) map[string][]FileInfo
}

type indexer struct {
	FileIndex map[string][]FileInfo
	mu        sync.Mutex // Mutex to protect shared resources
}

func NewIndexer() Indexer {
	return &indexer{
		FileIndex: make(map[string][]FileInfo),
	}
}

func (i *indexer) buildFileIndex(rootPath string) error {
	var wg sync.WaitGroup
	filePaths := make(chan string)

	// Start a pool of worker goroutines
	for j := 0; j < 20; j++ { // You can adjust the number of workers
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range filePaths {
				hash, err := i.fileHash(path)
				if err != nil {
					logconfig.Log.Errorf("Failed to hash file %s: %v", path, err)
					continue
				}
				fileName := filepath.Base(path)
				i.mu.Lock()
				i.FileIndex[fileName] = append(i.FileIndex[fileName], FileInfo{Path: path, Hash: hash})
				i.mu.Unlock()
			}
		}()
	}

	// Send file paths to be processed
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logconfig.Log.Errorf("The following error occurred %s", err)
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir // Skip dot directories
		}
		if !info.IsDir() {
			filePaths <- path
		}
		return nil
	})
	close(filePaths) // Close the channel to signal to the goroutines that there are no more files
	wg.Wait()        // Wait for all goroutines to finish

	return err
}

func (i *indexer) fileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logconfig.Log.Error(err)
		}
	}(file)
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (i *indexer) GetFileIndex(rootDir string) map[string][]FileInfo {
	if err := i.buildFileIndex(rootDir); err != nil {
		logconfig.Log.Errorf("the following error occurred %s", err)
		return nil
	}
	return i.FileIndex
}
