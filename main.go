package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

// IPResponse defines the structure of the JSON response
type IPResponse struct {
	IP string `json:"ip"`
}

var (
	currentIP string
	mutex     sync.RWMutex // Ensures safe concurrent access to currentIP
)

// getPublicIP fetches the public IP from icanhazip.com
func getPublicIP() (string, error) {
	resp, err := http.Get("https://icanhazip.com")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Return the trimmed IP (removing any newline characters)
	return string(body), nil
}

// refreshIP periodically updates the public IP every 30 minutes
func refreshIP() {
	for {
		ip, err := getPublicIP()
		if err != nil {
			log.Printf("Error refreshing IP: %v", err)
			time.Sleep(30 * time.Minute) // Retry after 30 minutes even if it fails
			continue
		}

		mutex.Lock()
		currentIP = ip
		mutex.Unlock()

		log.Printf("Public IP updated to: %s", currentIP)
		time.Sleep(30 * time.Minute)
	}
}

// ipHandler handles HTTP requests and responds with the current cached public IP in JSON
func ipHandler(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	ip := currentIP
	mutex.RUnlock()

	if ip == "" {
		http.Error(w, "Public IP not yet available. Please try again later.", http.StatusServiceUnavailable)
		return
	}

	response := IPResponse{IP: ip}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Start refreshing the IP in a separate goroutine
	go refreshIP()

	// Serve the IP via HTTP
	http.HandleFunc("/", ipHandler)

	log.Println("Server is running on port 6969...")
	err := http.ListenAndServe(":6969", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
