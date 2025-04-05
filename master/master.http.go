package master

import (
	"encoding/json"
	"fmt"
	"godrive/config"
	"log"
	"net/http"
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

	log.Println("\n\n>>>>> Listening to HTTP requests at:", port, " <<<<<\n")
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
		http.Error(w, "Couldn't parse json", http.StatusBadRequest)
		return
	}
	if incomingFile.Name == "" || incomingFile.Content == "" {
		http.Error(w, "fileName or content is empty", http.StatusBadRequest)
		return
	}
	BreakFilesIntoChunks(incomingFile)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Accepted file: %v", incomingFile.Name)))
}

func handleFileDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests allowed in this route", http.StatusBadRequest)
		return
	}
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "No key found", http.StatusBadRequest)
		return
	}
	chunkIndexMap, exists := metadata.Chunks[key]
	if !exists {
		return
	}
	getChunksFromSlaves(chunkIndexMap)
}

func getChunksFromSlaves(chunkIndexMap map[int]*ChunkInfo) {
	for ind := range chunkIndexMap {
		log.Println(ind)
	}
}
