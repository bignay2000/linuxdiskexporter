package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// DiskStat represents a single disk statistic.
type DiskStat struct {
	HostName string  `json:"hostname"`
	Path     string  `json:"path"`
	Size     int64   `json:"size"`
	Used     float64 `json:"used"`
	Type     string  `json:"type"`
	Read     float64 `json:"read"`
	Write    float64 `json:"write"`
}

// DiskStats represents a collection of disk statistics.
type DiskStats struct {
	Stats []DiskStat `json:"stats"`
}

func main() {
	// Create a new router
	router := mux.NewRouter()

	// Define the API endpoint for getting all disk statistics
	router.HandleFunc("/diskstats", getDiskStats).Methods("GET")
	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router, // Pass our instance of gorilla/mux in.
	}
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
}

// getDiskStats handles GET requests to the /diskstats endpoint.
func getDiskStats(w http.ResponseWriter, r *http.Request) {
	// Get the hostname from the request URL path
	var hostName string
	parts := r.URL.Path.Split("/")
	if len(parts) > 1 && parts[1] == "diskstats" {
		hostName = parts[0]
	}

	// Check if the hostname is valid
	if !isValidHostname(hostName) {
		http.Error(w, "Invalid hostname", http.StatusBadRequest)
		return
	}

	// Get the disk statistics for the given hostname
	stats := getDiskStatsForHostname(hostName)

	// Marshal the disk statistics to JSON
	json.NewEncoder(w).Encode(stats)
}

// isValidHostname checks if a given hostname is valid.
func isValidHostname(hostname string) bool {
	// This is a very basic implementation and may not cover all cases.
	// In a real-world application, you would want to use a more robust method to validate the hostname.
	return len(hostname) > 0
}

// getDiskStatsForHostname gets disk statistics for a given hostname.
func getDiskStatsForHostname(hostName string) ([]DiskStat, error) {
	// Run the df command and parse its output
	cmd := exec.Command("df", "-h", "/dev/"+hostName)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse the output to extract the disk statistics
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		size, _ := strconv.ParseInt(fields[1], 10, 64)
		used, _ := strconv.ParseFloat(fields[2], 64)
		read, _ := strconv.ParseFloat(fields[8], 64)
		write, _ := strconv.ParseFloat(fields[9], 64)

		stats := []DiskStat{
			{
				HostName: hostName,
				Path:     "/dev/" + hostName,
				Size:     size,
				Used:     used,
				Type:     fields[0],
				Read:     read,
				Write:    write,
			},
		}

		return stats, nil
	}

	return nil, fmt.Errorf("no disk statistics found for hostname %s", hostName)
}
