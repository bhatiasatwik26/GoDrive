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

func BreakFilesIntoChunks(incomingFile uploadedFile) {
	name, content := incomingFile.Name, incomingFile.Content
	// chunkSize := config.ReadConfig.Master.ChunkSize
	chunkSize := 4
	var createdFile FileStruct
	fileInBytes := []byte(content)
	createdFile.Name = name
	chunkInd := 0
	for i := 0; i < len(fileInBytes); i += chunkSize {
		end := min(len(fileInBytes), i+chunkSize)
		chunkHash := sha256.Sum256(fileInBytes[i:end])
		newChunk := FileChunk{
			Index: chunkInd,
			Data:  fileInBytes[i:end],
			Hash:  fmt.Sprintf("%x", chunkHash),
		}
		chunkInd += 1
		createdFile.Chunks = append(createdFile.Chunks, newChunk)
	}
	DistriButeChunksToNode(createdFile)
}
func MergeChunksToFile() uploadedFile {
	return uploadedFile{}
}
