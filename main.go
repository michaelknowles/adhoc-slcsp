package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
)

// File names
const SlcspFileName string = "slcsp.csv"
const ZipsFileName string = "zips.csv"
const PlansFileName string = "plans.csv"

// RateData holds the rating information for a zip code
// RateArea is a string where `state` and `rate_area` are concatenated from ZipsFileName/PlansFileName
// Rates is a slice of applicable rates found for the RateArea from PlansFileName
// Ambiguous marks whether a zip has multiple RateArea
type RateData struct {
	RateArea  string
	Rates     []float64
	Ambiguous bool
}

// concatRateArea creates the RateArea string for use in RateData
// It expects the `state` and the `rate_area` from ZipsFileName/PlansFileName
func concatRateArea(state string, code string) string {
	rateArea := fmt.Sprintf("%s%s", state, code)
	return rateArea
}

// parseSlcsp reads the data in SlcspFileName and returns all of the zip codes from it
func parseSlcsp() ([]string, error) {
	zips := make([]string, 0)
	slcspFile, err := os.Open(SlcspFileName)
	if err != nil {
		return zips, err
	}
	defer slcspFile.Close()

	slcspReader := csv.NewReader(slcspFile)
	slcspReader.FieldsPerRecord = 2

	// Skip first line (header)
	_, err = slcspReader.Read()
	if err != nil {
		return zips, err
	}

	// Read file data
	for {
		record, err := slcspReader.Read()

		// Stop at end of file
		if err == io.EOF {
			break
		}

		if err != nil {
			return zips, err
		}

		// Record fields:
		// 0 - zipcode
		// 1 - rate
		// Only store the zipcode field since rate will be empty here
		zips = append(zips, record[0])
	}

	return zips, err
}

// parseZips reads the data from ZipsFileName and adds RateArea info to the zip
func parseZips(zips map[string]*RateData) (map[string]*RateData, error) {
	zipsFile, err := os.Open(ZipsFileName)
	if err != nil {
		return zips, err
	}
	defer zipsFile.Close()

	zipsReader := csv.NewReader(zipsFile)
	zipsReader.FieldsPerRecord = 5

	// Skip first line (header)
	_, err = zipsReader.Read()
	if err != nil {
		return zips, err
	}

	// Read file data
	for {
		record, err := zipsReader.Read()

		// Stop at end of file
		if err == io.EOF {
			break
		}

		if err != nil {
			return zips, err
		}

		// Record fields:
		// 0 - zipcode
		// 1 - state
		// 2 - county_code
		// 3 - name
		// 4 - rate_area
		zip := record[0]
		// Store the rate area if the record's zipcode matches one in zips
		// If the rate area is already set and differs from the current record's mark the data as ambiguous
		if _, exists := zips[zip]; exists {
			rateArea := concatRateArea(record[1], record[4])
			if zips[zip].RateArea == "" {
				zips[zip].RateArea = rateArea
			} else if zips[zip].RateArea != rateArea {
				zips[zip].Ambiguous = true
			}
		}
	}

	return zips, err
}

// parsePlans reads the data from PlansFileName and adds Rates to the zip/RateArea
func parsePlans(zips map[string]*RateData) (map[string]*RateData, error) {
	plansFile, err := os.Open(PlansFileName)
	if err != nil {
		return zips, err
	}
	defer plansFile.Close()

	plansReader := csv.NewReader(plansFile)
	plansReader.FieldsPerRecord = 5

	// Skip first line (header)
	_, err = plansReader.Read()
	if err != nil {
		return zips, err
	}

	// Read file data
	for {
		record, err := plansReader.Read()

		// Stop at end of file
		if err == io.EOF {
			break
		}

		if err != nil {
			return zips, err
		}

		// Record fields:
		// 0 - plan_id
		// 1 - state
		// 2 - metal_level
		// 3 - rate
		// 4 - rate_area
		rateArea := concatRateArea(record[1], record[4])
		rate, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return zips, err
		}

		// Loop through each stored rate area
		// Store the rate if the record's rate area matches and it's a Silver plan
		// Skip the zip's rate area if it's been marked as ambiguous
		for _, rateData := range zips {
			if rateArea == rateData.RateArea && !rateData.Ambiguous && record[2] == "Silver" {
				rateData.Rates = append(rateData.Rates, rate)
			}
		}
	}

	return zips, err
}

func main() {
	// Read SlcspFileName
	zips, err := parseSlcsp()
	if err != nil {
		log.Fatal("Error parsing data from "+SlcspFileName, err)
	}

	// Create map from slice returned by parseSlcsp
	zipData := make(map[string]*RateData)
	for _, zip := range zips {
		zipData[zip] = &RateData{}
	}

	// Read ZipsFileName
	zipData, err = parseZips(zipData)
	if err != nil {
		log.Fatal("Error parsing data from "+ZipsFileName, err)
	}

	// Read PlansFileName
	zipData, err = parsePlans(zipData)
	if err != nil {
		log.Fatal("Error parsing data from "+PlansFileName, err)
	}

	// Output
	fmt.Println("zipcode,rate")
	for _, zip := range zips {
		rateData := zipData[zip]
		// If no second lowest rate, just output zip
		if len(rateData.Rates) < 2 {
			fmt.Println(zip + ",")
		} else {
			sort.Float64s(rateData.Rates) // sort least to greatest
			fmt.Println(fmt.Sprintf("%s,%.2f", zip, rateData.Rates[1]))
		}
	}
}
