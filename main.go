package main

import (
	"duplication-finder/indexer"
	"fmt"
)

func main() {
	NasIndexer := indexer.NewIndexer()
	fmt.Println("Indexing files...")

	// Retrieve and use the file index
	fileIndex := NasIndexer.GetFileIndex("/mnt/c/Users/chris/Documents/20051/Projects")
	fmt.Println("File indexing complete.")
	fmt.Println("Outputting map...")
	for fileName, files := range fileIndex {
		fmt.Printf("File: %s\n", fileName)
		for _, file := range files {
			fmt.Printf("\tPath: %s, Hash: %s\n", file.Path, file.Hash)
		}
	}
}
