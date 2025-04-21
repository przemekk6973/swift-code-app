// app/internal/util/countries.go
package util

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// LoadCountryMap wczytuje CSV z dwóch kolumn: ISO2,CountryName
func LoadCountryMap(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open countries file: %w", err)
	}
	defer f.Close()

	rdr := csv.NewReader(f)
	m := make(map[string]string)
	for {
		rec, err := rdr.Read()
		if err != nil {
			break
		}
		iso := strings.ToUpper(strings.TrimSpace(rec[0]))
		name := strings.ToUpper(strings.TrimSpace(rec[1]))
		m[iso] = name
	}
	return m, nil
}
