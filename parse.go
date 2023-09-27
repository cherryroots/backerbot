package main

import (
	"encoding/csv"
	"errors"
	"os"
	"strings"

	"github.com/rapidloop/skv"
)

type backer struct {
	Email    string
	Donation string
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
			Email:    record[1],
			Donation: record[2],
		}

		err := store.Put(b.Email, b)
		if err != nil {
			return err
		}
	}
	return nil
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
