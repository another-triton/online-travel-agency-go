package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Config struct {
	SupplierHostName string `json:"supplierHostName"`
	SupplierPort     string `json:"supplierPort"`
}

var config Config

func main() {
	fmt.Println("Starting adapter...")
	loadConfig()
	http.HandleFunc("/adapter/supplier", adapterHandler)
	if err := http.ListenAndServe(":9000", nil); err != nil {
		fmt.Printf("Failed to start adapter server: %s\n", err)
	}
}

func loadConfig() (Config, error) {

	configFile, err := os.Open("config.json")
	if err != nil {
		return config, err
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config, nil
}

func adapterHandler(w http.ResponseWriter, r *http.Request) {
	//call the supplier service which is hosted at 8090
	adapterUrl := "http://" + config.SupplierHostName + ":" + config.SupplierPort + "/api/supplier?supplierId=0"
	//fmt.Println("Calling supplier service at: ", adapterUrl)
	resp, err := http.Get(adapterUrl)
	if err != nil {
		fmt.Println("Failed to call supplier service")
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Adapter failed to read response body")
	}
	w.Write(body)
}
