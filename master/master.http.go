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

func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func StartMasterHttp() {
	port := config.ReadConfig.Master.HttpPort
	fullAddress := fmt.Sprintf(":%d", port)
	MyNodeSelector = NewRoundRobinSelector(config.ReadConfig.SlaveNodes)

	http.HandleFunc("/", healthCheck)
	http.HandleFunc("/upload", handleFileUpload)
	http.HandleFunc("/download", handleFileDownload)
	http.HandleFunc("/update", handleFileUpdate)
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
func DeleteFileIfQuoramFails(filename string) {

	metadata.mu.Lock()
	fileInfo, exists := metadata.Chunks[filename]
	metadata.mu.Unlock()

	if !exists {
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
	}
}

// Route handler functions
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Http server looks good"))
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {

	enableCORS(w, r)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	if _, exists := metadata.Chunks[incomingFile.Name]; exists {
		http.Error(w, "File already present in system. Delete file first to use upload or use 'update'.", http.StatusConflict)
		return
	}
	createdFile := BreakFilesIntoChunks(incomingFile)
	if success := DistriButeChunksToNode(createdFile); success {
		log.Println("────────────────────────────────────────")
		log.Printf("✅  Master: Splitted file into %v chunks", len(createdFile.Chunks))
		log.Println("────────────────────────────────────────")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Accepted file: '%v'.", incomingFile.Name)))
	} else {
		log.Println("────────────────────────────────────────")
		log.Printf("⚠️  Master: Failed to upload file")
		log.Println("────────────────────────────────────────")
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
	}

}

func handleFileDownload(w http.ResponseWriter, r *http.Request) {

	enableCORS(w, r)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

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
		if incomingFileChunk.Index == -1 {
			log.Println("────────────────────────────────────────")
			log.Printf("⚠️  Master: Error while downloading file")
			log.Println("────────────────────────────────────────")
			http.Error(w, "Error in Downloading file", http.StatusInternalServerError)
			return
		}
		downloadedFile.Chunks[incomingFileChunk.Index] = incomingFileChunk
	}
	log.Println("────────────────────────────────────────")
	log.Printf("✅ Master: Download %v sucessfully\n", downloadedFile.Name)
	log.Println("────────────────────────────────────────")
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

func handleFileDelete(w http.ResponseWriter, r *http.Request) {

	enableCORS(w, r)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

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
		fmt.Fprintf(w, "Deleted '%s' from the system.\n", filename)
	} else {
		log.Println("Couldn't delete file")
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
	}
}

func handleFileUpdate(w http.ResponseWriter, r *http.Request) {

	enableCORS(w, r)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var incoming uploadedFile
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil || incoming.Name == "" || incoming.Content == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	newFile := BreakFilesIntoChunks(incoming)

	metadata.mu.Lock()
	existingChunks, exists := metadata.Chunks[incoming.Name]
	metadata.mu.Unlock()

	if !exists {
		http.Error(w, "File does not exist. Upload it first.", http.StatusNotFound)
		return
	}

	// Compare chunk count (optional step)
	if len(newFile.Chunks) != len(existingChunks) {
		log.Printf("Chunk count changed from %d to %d for file %s\n", len(existingChunks), len(newFile.Chunks), incoming.Name)
	}

	updatedChunks := make(map[int]*ChunkInfo)
	var deleteWg sync.WaitGroup
	deleteAckChan := make(chan bool)

	for _, newChunk := range newFile.Chunks {
		oldChunk, found := existingChunks[newChunk.Index]

		if found && oldChunk.ChunkHash == newChunk.Hash {
			// Chunk unchanged, keep old metadata
			updatedChunks[newChunk.Index] = oldChunk
			continue
		}

		// Chunk changed: delete old chunk if exists
		if found && oldChunk != nil {
			deleteWg.Add(1)
			go deleteChunkFromSlaves(oldChunk, deleteAckChan, &deleteWg)
		}

		// Assign new nodes and distribute chunk
		success, nodePorts := assignNode(newChunk)
		if !success {
			http.Error(w, "Failed to distribute chunk to slave nodes", http.StatusInternalServerError)
			return
		}

		updatedChunks[newChunk.Index] = &ChunkInfo{
			ChunkHash:     newChunk.Hash,
			SlaveNodeList: nodePorts,
		}
	}

	// Wait for all deletions to complete
	go func() {
		deleteWg.Wait()
		close(deleteAckChan)
	}()
	for range deleteAckChan {
	}

	metadata.mu.Lock()
	metadata.Chunks[incoming.Name] = updatedChunks
	metadata.mu.Unlock()
	SaveMetaDataToFile()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File '%s' updated successfully\n", incoming.Name)
}
