package master

import "log"

type FileChunk struct {
	Index int
	Data  []byte
	Hash  []byte
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
		newChunk := FileChunk{
			Index: chunkInd,
			Data:  fileInBytes[i:end],
		}
		chunkInd += 1
		createdFile.Chunks = append(createdFile.Chunks, newChunk)
	}
	log.Println(len(createdFile.Chunks))
	DistriButeChunksToNode(createdFile)
}
