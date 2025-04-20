package master

import (
	"crypto/sha256"
	"fmt"
)

type FileChunk struct {
	Index int    `json:"index"`
	Data  []byte `json:"data"`
	Hash  string `json:"hash"`
}

type FileStruct struct {
	Name   string
	Chunks []FileChunk
}

func BreakFilesIntoChunks(incomingFile uploadedFile) FileStruct {
	name, content := incomingFile.Name, incomingFile.Content
	// chunkSize := config.ReadConfig.Master.ChunkSize
	chunkSize := 4
	var createdFile FileStruct
	contentInBytes := []byte(content)
	createdFile.Name = name
	chunkInd := 0
	for i := 0; i < len(contentInBytes); i += chunkSize {
		end := min(len(contentInBytes), i+chunkSize)
		chunkHash := sha256.Sum256(contentInBytes[i:end])
		newChunk := FileChunk{
			Index: chunkInd,
			Data:  contentInBytes[i:end],
			Hash:  fmt.Sprintf("%x", chunkHash),
		}
		chunkInd += 1
		createdFile.Chunks = append(createdFile.Chunks, newChunk)
	}
	return createdFile
}
func MergeChunksToFile(downloadedFile FileStruct) uploadedFile {
	createdFile := uploadedFile{
		Name: downloadedFile.Name,
	}
	var contentInBytes []byte
	for _, data := range downloadedFile.Chunks {
		contentInBytes = append(contentInBytes, data.Data...)
	}
	createdFile.Content = string(contentInBytes)
	return createdFile
}
