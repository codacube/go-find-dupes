package exporter

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"

	"github.com/codacube/go-find-dupes/internal/model"
)

func ToJSON(groups []model.FileGroup, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(groups)
}

func ToCSV(groups []model.FileGroup, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// TODO Finish this, export the columns we actually want
	header := []string{"Group Hash", "File Path", "Size (Bytes)", "Is Original?"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, group := range groups {
		for i, f := range group.Files {
			isOriginal := "No"
			if i == 0 {
				isOriginal = "Yes"
			}

			row := []string{
				group.Identifier, // The Hash
				f.DisplayPath,    // The Location
				strconv.FormatInt(f.Size, 10),
				isOriginal,
			}

			if err := writer.Write(row); err != nil {
				return err
			}
		}
	}

	return nil
}
