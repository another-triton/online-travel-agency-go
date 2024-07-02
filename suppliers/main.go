package main

import (
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var filesContent []string

func convertStrToInt(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return num
}

func sourceHandler(w http.ResponseWriter, r *http.Request) {

	//currentTime := time.Now()
	maxSleepSimulationTime := 4

	// Retrieve the integer value from the query parameter
	integerStr := r.URL.Query().Get("supplierId")

	// Check if the integer value is provided
	if integerStr == "" {
		http.Error(w, "Integer value is missing", http.StatusBadRequest)
		return
	}

	// convert string to integer and return 0 if input string is not a valid integer

	integer := convertStrToInt(integerStr)

	//get a random number between 1 and 5 without any seed
	var randomInteger = rand.Intn(maxSleepSimulationTime) + 1

	// Sleep for the duration of the integer value
	time.Sleep(time.Duration(randomInteger) * time.Second)

	//fmt.Println("reading file content at index: ", integer)
	// Get the file content at the index of the integer value
	content, err := getFileContent(integer)

	if err != nil {
		http.Error(w, "Error getting file content", http.StatusInternalServerError)
		return
	}

	//send the var content which is a output of getFileContent as HTTP response
	fmt.Fprintf(w, "%s", content)
	//fmt.Println("Time taken by the supplier: ", time.Since(currentTime))

}

func loadAllFiles() {
	directory := "./xmls"
	// Read all XML files from the directory and store their content in memory
	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		fmt.Printf("Reading file: %s %s \n", path, d.Name())
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Check if the file has a .xml extension
		if filepath.Ext(d.Name()) == ".xml" {
			// Read the file content
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Store the file content in the slice
			filesContent = append(filesContent, string(content))
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func getFileContent(index int) (string, error) {
	// Check if the index is valid
	if index < 0 || index >= len(filesContent) {
		return "", fmt.Errorf("index out of range")
	}

	// Return the file content at the given index
	return filesContent[index], nil
}

func main() {
	fmt.Println("Loading all files...")
	loadAllFiles()
	// Define the route for the API endpoint
	http.HandleFunc("/api/supplier", sourceHandler)

	// Start the HTTP server on port 8080
	fmt.Println("Server starting on port 8080")
	fmt.Println("Endpoint is : http://localhost:8080/api/supplier?supplierId=0")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
	}
}
