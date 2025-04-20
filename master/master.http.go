package master

import (
	"encoding/json"
	"fmt"
	"godrive/config"
	"log"
	"net/http"
	"sync"
)

type uploadedFile struct {
	Name    string `json:"fileName"`
	Content string `json:"content"`
}

var MyNodeSelector *RoundRobinNodeSelector

func StartMasterHttp() {
	port := config.ReadConfig.Master.HttpPort
	fullAddress := fmt.Sprintf(":%d", port)
	MyNodeSelector = NewRoundRobinSelector(config.ReadConfig.SlaveNodes)

	http.HandleFunc("/", healthCheck)
	http.HandleFunc("/upload", handleFileUpload)
	http.HandleFunc("/download", handleFileDownload)
	// http.HandleFunc("/update", handleFileUpdate)
	http.HandleFunc("/delete", handleFileDelete)

	log.Printf(`
╔════════════════════════════════════════╗
║   HTTP SERVER STARTED ON PORT %v     ║
╚════════════════════════════════════════╝`, port)
	err := http.ListenAndServe(fullAddress, nil)
	if err != nil {
		log.Fatal("HTTP server crashed")
	}
}

// Route handler functions
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Http server looks good"))
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests allowed in this route", http.StatusBadRequest)
		return
	}
	var incomingFile uploadedFile
	err := json.NewDecoder(r.Body).Decode(&incomingFile)
	if err != nil {
		http.Error(w, "Bad format file", http.StatusBadRequest)
		return
	}
	if incomingFile.Name == "" || incomingFile.Content == "" {
		http.Error(w, "FileName or content is empty", http.StatusBadRequest)
		return
	}
	createdFile := BreakFilesIntoChunks(incomingFile)
	DistriButeChunksToNode(createdFile)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Accepted file: %v", incomingFile.Name)))
}

func handleFileDownload(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests allowed in this route", http.StatusBadRequest)
		return
	}
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "No key found", http.StatusBadRequest)
		return
	}
	indexToFilechunkMap, exists := metadata.Chunks[filename]
	if !exists {
		http.Error(w, "No such file present in the system.", http.StatusNotFound)
		return
	}
	downloadedFile := FileStruct{
		Name:   filename,
		Chunks: make([]FileChunk, len(indexToFilechunkMap)),
	}
	incomingChunksChannel := make(chan FileChunk)
	var wg sync.WaitGroup
	for index, chunkInfo := range indexToFilechunkMap {
		wg.Add(1)
		go getChunk(index, chunkInfo, incomingChunksChannel, &wg)
	}
	go func() {
		wg.Wait()
		close(incomingChunksChannel)
	}()
	for incomingFileChunk := range incomingChunksChannel {
		downloadedFile.Chunks[incomingFileChunk.Index] = incomingFileChunk
	}
	log.Printf("⬇️  Download %v sucessfully\n", downloadedFile.Name)
	createdFileAfterMerge := MergeChunksToFile(downloadedFile)
	createdFileJson, err := json.Marshal(createdFileAfterMerge)
	if err != nil {
		log.Println("Error")
		return
	}
	w.Write(createdFileJson)
}
func getChunk(index int, chunkInfo *ChunkInfo, channelToSendChunk chan FileChunk, wg *sync.WaitGroup) {
	defer wg.Done()
	chunkHash, slaveNodeList := chunkInfo.ChunkHash, chunkInfo.SlaveNodeList
	for ind := 0; ind < len(slaveNodeList); ind++ {
		obtainedFileChunk, err := RequestChunkFromSlave(slaveNodeList[ind], chunkHash)
		if err == nil {
			obtainedFileChunk.Index = index
			channelToSendChunk <- obtainedFileChunk
			return
		}
	}
	channelToSendChunk <- FileChunk{Index: -1}
}

// func handleFileUpdate(w http.ResponseWriter, r *http.Request) {

// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Only POST requests allowed in this route", http.StatusBadRequest)
// 		return
// 	}
// 	var incomingFile uploadedFile
// 	err := json.NewDecoder(r.Body).Decode(&incomingFile)
// 	if err != nil {
// 		http.Error(w, "Bad format file", http.StatusBadRequest)
// 		return
// 	}
// 	if incomingFile.Name == "" || incomingFile.Content == "" {
// 		http.Error(w, "FileName or content is empty", http.StatusBadRequest)
// 		return
// 	}
// 	if metadata.Chunks[incomingFile.Name] == nil {
// 		http.Error(w, "No such file found to update", http.StatusNotFound)
// 		return
// 	}
// 	createdFile := BreakFilesIntoChunks(incomingFile)
// 	CompareChunksAndUpdate(createdFile)
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(fmt.Sprintf("Accepted file: %v", incomingFile.Name)))
// }

func handleFileDelete(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE requests are allowed on this route", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "No filename provided", http.StatusBadRequest)
		return
	}

	metadata.mu.Lock()
	fileInfo, exists := metadata.Chunks[filename]
	metadata.mu.Unlock()

	if !exists {
		http.Error(w, "No such file present in the system.", http.StatusNotFound)
		return
	}

	deleteAckChannel := make(chan bool)
	var wg sync.WaitGroup

	for _, chunkInfo := range fileInfo {
		wg.Add(1)
		go deleteChunkFromSlaves(chunkInfo, deleteAckChannel, &wg)
	}

	go func() {
		wg.Wait()
		close(deleteAckChannel)
	}()

	allSuccessful := true
	for ack := range deleteAckChannel {
		if !ack {
			allSuccessful = false
			break
		}
	}

	if allSuccessful {
		metadata.mu.Lock()
		delete(metadata.Chunks, filename)
		metadata.mu.Unlock()
		SaveMetaDataToFile()
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, " Deleted '%s' from the system.\n", filename)
	} else {
		log.Println("Couldn't delete file")
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
	}
}
