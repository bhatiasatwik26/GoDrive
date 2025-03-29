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

	log.Println("Listening to HTTP requests at", fullAddress)
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
