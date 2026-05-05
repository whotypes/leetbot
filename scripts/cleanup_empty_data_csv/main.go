// Command cleanup_empty_data_csv removes CSV files under data/ that have no data rows
// (header-only or empty). Company directories are left in place. Run from repo root:
//
//	go run ./scripts/cleanup_empty_data_csv
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dataDir := filepath.Join(cwd, "data")
	var removed, kept int
	err = filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".csv") {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		rows, err := readAllRows(f)
		if err != nil {
			return err
		}
		// No header+data, or only empty second row
		if len(rows) <= 1 {
			if err := os.Remove(path); err != nil {
				return err
			}
			fmt.Println("removed", path)
			removed++
			return nil
		}
		if len(rows) == 2 && rowAllBlank(rows[1]) {
			if err := os.Remove(path); err != nil {
				return err
			}
			fmt.Println("removed", path)
			removed++
			return nil
		}
		kept++
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("done: removed %d, kept %d (with at least one data row)\n", removed, kept)
}

func readAllRows(r io.Reader) ([][]string, error) {
	return csv.NewReader(r).ReadAll()
}

func rowAllBlank(cols []string) bool {
	for _, c := range cols {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}
