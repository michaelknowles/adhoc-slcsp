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

const SlcspFileName string = "slcsp.csv"
const ZipsFileName string = "zips.csv"
const PlansFileName string = "plans.csv"

type RateData struct {
	RateArea string
	Rates    []float64
}

func concatRateArea(state string, code string) string {
	rateArea := fmt.Sprintf("%s %s", state, code)
	return rateArea
}

func parseSlcsp() (map[string]*RateData, error) {
	zips := make(map[string]*RateData)
	slcspFile, err := os.Open(SlcspFileName)
	if err != nil {
		return zips, err
	}
	defer slcspFile.Close()

	slcspReader := csv.NewReader(slcspFile)
	slcspReader.FieldsPerRecord = 2

	// Skip first line
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
		zips[record[0]] = &RateData{}
	}

	return zips, err
}

func parseZips(zips map[string]*RateData) (map[string]*RateData, error) {
	zipsFile, err := os.Open(ZipsFileName)
	if err != nil {
		return zips, err
	}
	defer zipsFile.Close()

	zipsReader := csv.NewReader(zipsFile)
	zipsReader.FieldsPerRecord = 5

	// Skip first line
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
		if _, exists := zips[zip]; exists {
			rateArea := concatRateArea(record[1], record[4])
			zips[zip].RateArea = rateArea
		}
	}

	return zips, err
}

func parsePlans(zips map[string]*RateData) (map[string]*RateData, error) {
	plansFile, err := os.Open(PlansFileName)
	if err != nil {
		return zips, err
	}
	defer plansFile.Close()

	plansReader := csv.NewReader(plansFile)
	plansReader.FieldsPerRecord = 5

	// Skip first line
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

		for _, rateData := range zips {
			if rateArea == rateData.RateArea && record[2] == "Silver" {
				rateData.Rates = append(rateData.Rates, rate)
			}
		}

	}

	return zips, err
}

func main() {
	// Read slcsp.csv
	zips, err := parseSlcsp()
	if err != nil {
		log.Fatal("Error parsing data from "+SlcspFileName, err)
	}

	// Read zips.csv
	zips, err = parseZips(zips)
	if err != nil {
		log.Fatal("Error parsing data from "+ZipsFileName, err)
	}

	// Read plans.csv
	zips, err = parsePlans(zips)
	if err != nil {
		log.Fatal("Error parsing data from "+PlansFileName, err)
	}

	// Output
	fmt.Println("zipcode,rate")
	for zip, rateData := range zips {
		// If no second lowest rate, just output zip
		if len(rateData.Rates) < 2 {
			fmt.Println(zip + ",")
		} else {
			// Get second lowest rate
			sort.Float64s(rateData.Rates)
			fmt.Println(fmt.Sprintf("%s,%.2f", zip, rateData.Rates[1]))
		}
	}
}
