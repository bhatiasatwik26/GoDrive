package master

type FileChuks struct {
	Index int
	Data  []byte
	Hash  []byte
}
type FileStruct struct {
	FileName string

	FileChunks [][]byte
}
