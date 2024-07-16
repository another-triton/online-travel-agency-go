package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Define XML structures
type HotelFindResponse struct {
	XMLName          xml.Name `xml:"HotelFindResponse"`
	Time             string   `xml:"time,attr"`
	IPAddress        string   `xml:"ipaddress,attr"`
	Count            int      `xml:"count,attr"`
	ArrivalDate      string   `xml:"ArrivalDate"`
	DepartureDate    string   `xml:"DepartureDate"`
	Currency         string   `xml:"Currency"`
	GuestNationality string   `xml:"GuestNationality"`
	SearchSessionId  string   `xml:"SearchSessionId"`
	Hotels           []Hotel  `xml:"Hotels>Hotel"`
}

type Hotel struct {
	XMLName   xml.Name `xml:"Hotel"`
	XMLString string   `xml:",innerxml"`
}

type HotelsSlice struct {
	Hotels []Hotel `xml:"Hotel"` // Adjust the XML tag as necessary based on the actual XML structure
}

type ProcessorSettings struct {
	CPUUsageInMilliseconds struct {
		Min int
		Max int
	}
}
type Config struct {
	MinNoOfSuppliersForRandomness int    `json:"minNoOfSuppliersForRandomness"`
	MaxNoOfSuppliersForRandomness int    `json:"maxNoOfSuppliersForRandomness"`
	MinCpuUsageInMilliseconds     int    `json:"minCpuUsageInMilliseconds"`
	MaxCpuUsageInMilliseconds     int    `json:"maxCpuUsageInMilliseconds"`
	AdapterHostName               string `json:"adapterHostName"`
	AdapterPort                   string `json:"adapterPort"`
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

var config Config

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if false {
		// 	fmt.Println("skip gzip middleware")
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fmt.Println("client does not support gzip encoding")
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzrw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzrw, r)
	})
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
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

var minNoOfSuppliersForRandomness = 2
var maxNoOfSuppliersForRandomness = 5
var adapterHostUrl = ""
var minCpuUsageForSimulation = 500
var maxCpuUsageForSimulation = 1000

func main() {
	fmt.Println("number of current go procs", runtime.GOMAXPROCS(0))
	//runtime.GOMAXPROCS(12)
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
		return
	}
	adapterHostUrl = "http://" + config.AdapterHostName + ":" + config.AdapterPort + "/adapter/supplier?supplierId="
	minNoOfSuppliersForRandomness = config.MinNoOfSuppliersForRandomness
	maxNoOfSuppliersForRandomness = config.MaxNoOfSuppliersForRandomness

	minCpuUsageForSimulation = config.MinCpuUsageInMilliseconds
	maxCpuUsageForSimulation = config.MaxCpuUsageInMilliseconds

	// Wrap your existing handler with the gzip middleware
	//compressedHandler := gzipMiddleware(http.HandlerFunc(getAccomodationHandler))
	//http.Handle("/get-accomodations/{id}", compressedHandler)

	http.HandleFunc("/get-accomodations/{id}", getAccomodationHandler)

	fmt.Println("starting server at :8090")

	if err := http.ListenAndServe(":8090", nil); err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
	}

}

func getAccomodationHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL parameters
	//currentTime := time.Now()
	//fmt.Println("get accomodation handler called")
	result, err := GetAccomodations()
	if err != nil {
		http.Error(w, "Error getting accomodations", http.StatusInternalServerError)
		return
	}
	//fmt.Println("Time taken to get accomodations: ", time.Since(currentTime))
	// For now, just write the ID back to the client
	fmt.Fprintf(w, "%s", result)
}

func GetAccomodationBySupplierAsync(supplierId int) (string, error) {
	url := adapterHostUrl + strconv.Itoa(supplierId)

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

func randomBetween(min, max int) int {
	// if min >= max {
	// 	return 0, fmt.Errorf("invalid range: %d >= %d", min, max)
	// }
	n := rand.Intn(max - min)
	return n + min
}

func GetAccomodations() (string, error) {
	supplierList := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	numberOfSuppliers := randomBetween(minNoOfSuppliersForRandomness, maxNoOfSuppliersForRandomness)
	supplierFromIndex := rand.Intn(len(supplierList))

	suppliers := make([]int, numberOfSuppliers)
	for i := 0; i < numberOfSuppliers; i++ {
		suppliers[i] = supplierList[supplierFromIndex]
		supplierFromIndex = (supplierFromIndex + 1) % len(supplierList)
	}

	var wg sync.WaitGroup
	results := make([]string, len(suppliers))
	for i, supplier := range suppliers {
		wg.Add(1)
		go func(i, supplier int) {
			defer wg.Done()
			//currentTime := time.Now()
			res, err := GetAccomodationBySupplierAsync(supplier)
			//fmt.Println("Time taken to get accomodation from supplier: ", supplier, time.Since(currentTime))
			if err != nil {
				fmt.Println(err)
				return
			}
			results[i] = res
		}(i, supplier)
	}
	wg.Wait()

	//hotelCount := 0
	xmlTemplate := "<HotelFindResponse time=\"0.21500015258789\" ipaddress=\"14.140.153.130\" count=\"0\">\r\n    <ArrivalDate>01/06/2024</ArrivalDate>\r\n    <DepartureDate>10/06/2024</DepartureDate>\r\n    <Currency>INR</Currency>\r\n    <GuestNationality>IN</GuestNationality>\r\n    <SearchSessionId>17168872488751716887248949665</SearchSessionId><Hotels></Hotels></HotelFindResponse>"
	response := HotelFindResponse{}
	err := xml.Unmarshal([]byte(xmlTemplate), &response)
	if err != nil {
		fmt.Println("Error unmarshalling XML template:", err)
		return "", err
	}
	for _, result := range results {
		hotels := HotelsSlice{}
		err := xml.Unmarshal([]byte(result), &hotels)
		if err != nil {
			fmt.Println("Error unmarshalling result:", err)
			continue
		}
		//fmt.Println("hotels: ", hotels)
		response.Hotels = append(response.Hotels, hotels.Hotels...)
	}

	response.Count = len(response.Hotels)

	output, err := xml.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling XML:", err)
		return "", err
	}

	//fmt.Println("finalResult: ", finalResult)
	simulateCpuUsage(&results[0])

	finalResult := string(output)
	return finalResult, nil
}

func simulateCpuUsage(xmlDocument *string) {

	loopTillTime := time.Now().Add(
		time.Duration(randomBetween(minCpuUsageForSimulation, maxCpuUsageForSimulation)) * time.Millisecond)

	loopCounter := 0
	if xmlDocument != nil {
		for time.Now().Before(loopTillTime) {
			//mergedDoc = append(mergedDoc, *xmlDocument)
			//get the hash code of the xmlDocument
			hashXMLDocument(xmlDocument)
			loopCounter++
		}
	}
	//fmt.Println("loopCounter: ", loopCounter)

	//mergedDoc = nil
	//xmlDocument = nil
}

func hashXMLDocument(xmlDocument *string) string {
	if xmlDocument == nil {
		return ""
	}
	//now := time.Now()
	// Calculate the number of 100-nanosecond intervals since January 1, 1970 (Unix epoch)
	//ticksSinceEpoch := now.UnixNano() / 100
	// Calculate the number of 100-nanosecond intervals between January 1, 0001, and January 1, 1970
	// 621355968000000000 is the number of ticks from January 1, 0001 to January 1, 1970
	//ticks := ticksSinceEpoch + 621355968000000000
	text := *xmlDocument + createNewGUID()
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func createNewGUID() string {
	newGUID := uuid.New()
	return newGUID.String()
}
