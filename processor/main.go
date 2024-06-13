package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Hotel struct {
	XMLName  xml.Name `xml:"Hotel"`
	Contents string   `xml:",innerxml"`
}
type Hotels struct {
	XMLName xml.Name `xml:"Hotels"`
	Hotels  []Hotel  `xml:"Hotel"`
}
type Root struct {
	XMLName xml.Name `xml:"root"`
	Hotels  Hotels   `xml:"Hotels"`
}
type ProcessorSettings struct {
	CPUUsageInMilliseconds struct {
		Min int
		Max int
	}
}
type Config struct {
	MaxNoOfSuppliersForRandomness int    `json:"maxNoOfSuppliersForRandomness"`
	MinCpuUsageInMilliseconds     int    `json:"minCpuUsageInMilliseconds"`
	MaxCpuUsageInMilliseconds     int    `json:"maxCpuUsageInMilliseconds"`
	SupplierHostName              string `json:"supplierHostName"`
	SupplierPort                  string `json:"supplierPort"`
}

var config Config

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

var maxNoOfSuppliersForRandomness = 5
var supplierHostUrl = ""

func main() {

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
		return
	}
	supplierHostUrl = "http://" + config.SupplierHostName + ":" + config.SupplierPort + "/api/supplier?supplierId="
	maxNoOfSuppliersForRandomness = config.MaxNoOfSuppliersForRandomness

	http.HandleFunc("/get-accomodations/{id}", getAccomodationHandler)
	fmt.Println("starting server at :8090")

	if err := http.ListenAndServe(":8090", nil); err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
	}

}

func getAccomodationHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL parameters

	result, err := GetAccomodations()
	if err != nil {
		http.Error(w, "Error getting accomodations", http.StatusInternalServerError)
		return
	}
	// For now, just write the ID back to the client
	fmt.Fprintf(w, "%s", result)
}

// type ProcessorSettings struct {
// 	MaxNoOfSupplier int
// }

// type SupplierClient interface {
// 	GetAccomodationBySupplierAsync(int) (string, error)
// }

func GetAccomodationBySupplierAsync(supplierId int) (string, error) {
	url := +strconv.Itoa(supplierId)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func GetAccomodations() (string, error) {
	supplierList := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	maxNoOfSupplier := rand.Intn(maxNoOfSuppliersForRandomness) + 1
	supplierFromIndex := rand.Intn(len(supplierList))

	suppliers := make([]int, maxNoOfSupplier)
	for i := 0; i < maxNoOfSupplier; i++ {
		suppliers[i] = supplierList[supplierFromIndex]
		supplierFromIndex = (supplierFromIndex + 1) % len(supplierList)
	}

	var wg sync.WaitGroup
	results := make([]string, len(suppliers))
	for i, supplier := range suppliers {
		wg.Add(1)
		go func(i, supplier int) {
			defer wg.Done()
			res, err := GetAccomodationBySupplierAsync(supplier)
			if err != nil {
				fmt.Println(err)
				return
			}
			results[i] = res
		}(i, supplier)
	}
	wg.Wait()

	hotelCount := 0
	var result strings.Builder
	result.WriteString("<HotelFindResponse time=\"0.21500015258789\" ipaddress=\"14.140.153.130\" count=\"0\">\r\n    <ArrivalDate>01/06/2024</ArrivalDate>\r\n    <DepartureDate>10/06/2024</DepartureDate>\r\n    <Currency>INR</Currency>\r\n    <GuestNationality>IN</GuestNationality>\r\n    <SearchSessionId>17168872488751716887248949665</SearchSessionId><Hotels>")
	for _, xmlStr := range results {
		var root Root
		err := xml.Unmarshal([]byte("<root>"+xmlStr+"</root>"), &root)
		if err != nil {
			fmt.Println(err)
			return "", err
		}
		hotelCount += len(root.Hotels.Hotels)
		fmt.Println("hotel count: ", hotelCount)
		for _, hotel := range root.Hotels.Hotels {
			hotelXML, err := xml.Marshal(hotel)
			if err != nil {
				fmt.Println(err)
				return "", err
			}
			//fmt.Println("hotel xml is: ", string(hotelXML))
			result.WriteString(string(hotelXML))
		}
	}
	result.WriteString("\r\n</Hotels>\r\n</HotelFindResponse>")
	finalResult := replaceAtIndex(result.String(), strconv.Itoa(hotelCount), 60)
	simulateCpuUsage(&finalResult)
	return finalResult, nil
}

func replaceAtIndex(in, value string, i int) string {
	out := []rune(in)
	for j := 0; j < len(value); j++ {
		out[i+j] = rune(value[j])
	}
	return string(out)
}

func simulateCpuUsage(xmlDocument *string) {

	var mergedDoc []string
	minCpuUsageInMilliseconds := config.MinCpuUsageInMilliseconds
	maxCpuUsageInMilliseconds := config.MaxCpuUsageInMilliseconds

	//rand.Seed(time.Now().UnixNano())
	loopTillTime := time.Now().Add(time.Duration(rand.Intn(maxCpuUsageInMilliseconds-minCpuUsageInMilliseconds)+minCpuUsageInMilliseconds) * time.Millisecond)

	if xmlDocument != nil {
		for time.Now().Before(loopTillTime) {
			mergedDoc = append(mergedDoc, *xmlDocument)
		}
	}

	mergedDoc = nil
	xmlDocument = nil
}
