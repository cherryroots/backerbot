package main

import (
	"backerbot/skv"
	"encoding/csv"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type backer struct {
	Email       string
	RewardTitle string
	Donation    float64
	Status      string
}

func parse(csvString string) error {
	os.Remove("backers.db")

	store, err := skv.Open("backers.db")
	if err != nil {
		return err
	}
	defer store.Close()

	records, err := readData(csvString)
	if err != nil {
		return errors.New("failed to parse csv")
	}

	for _, record := range records {
		b := backer{
			Email:       strings.ToLower(record[0]),
			RewardTitle: record[1],
			Donation:    cleanDonation(record[2]),
			Status:      record[3],
		}

		err := store.Put(b.Email, b)
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanDonation(donation string) float64 {
	// regex [^\d.]+
	var re = regexp.MustCompile(`[^\d.]+`)

	donation = re.ReplaceAllString(donation, "")

	// convert string to float
	donationFloat, err := strconv.ParseFloat(donation, 64)
	if err != nil {
		return 0
	}

	return donationFloat
}

func readData(csvString string) ([][]string, error) {
	r := csv.NewReader(strings.NewReader(csvString))

	// skip header of the csv
	if _, err := r.Read(); err != nil {
		return [][]string{}, err
	}

	records, err := r.ReadAll()

	if err != nil {
		return [][]string{}, err
	}

	return records, nil
}
