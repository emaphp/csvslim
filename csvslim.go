// csvslim is a small utility for manipulating CSV files.

// Build Instructions:
//   go get "github.com/alexflint/go-arg"
//   go get "robpike.io/filter"
//   go build csvslim.go

// Usage:
//   ./csvslim -c [COLUMN1,COLUMN2,...] < input.csv

// Where COLUMN is the number corresponding to that column (starting at 0).
// It can also include a comparison operator (> and <).

// Examples:

//  Show second and fourth columns:
//   ./csvslim -c "1,3" < slim.csv

//  Show the first 3 columns and the fifth (0, 1, 2, 4)
//   ./csvslim -c "<3,4" < input.csv

// Ignore colums:
//   ./csvslim -i [COLUMN1,COLUMN2,...] < input.csv

// Rename columns (rename columns by index):
//   ./csvslim -r "COLUMN1:RENAME1,COLUMN2:RENAME2,..." < input.csv

// Ignore header (skip first line):
//   ./csvslim --noheader < input.csv

// Filter by value (filter file must contain sorted values):
//   ./csvslim --filter filter.csv < input.csv

// Filter by value specifying the column to watch:
//   ./csvslim --filter filter.csv --filtercol 1 < input.csv

// Inverse filter:
//   ./csvslim --filter filter.csv --filtercol 1 --inverse < input.csv

package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/alexflint/go-arg"
	"io"
	"log"
	"os"
	"regexp"
	"robpike.io/filter"
	"strconv"
	"strings"
)

// The Operator type identifies a comparison operator to be used againts a column number
type Operator string

const (
	Equal       Operator = "="
	LessThan    Operator = "<"
	GreaterThan Operator = ">"
)

// A ColumnOperator stores a comparison operator along with the column number it needs to be compared with
type ColumnOperator struct {
	Column     int
	Comparison Operator
}

// The Evaluate method compares a column number against the column within the ColumnOperator
func (ce *ColumnOperator) Evaluate(c int) bool {
	switch ce.Comparison {
	case Equal:
		return c == ce.Column
	case LessThan:
		return c < ce.Column
	case GreaterThan:
		return c > ce.Column
	}

	return false
}

// The FilterColumns type holds all ColumnOperators values a column number must be compared with
type FilterColumns struct {
	Values []ColumnOperator
}

// The IsValid method checks whether the column number passes all checks within a FilterColumns instance
func (fc *FilterColumns) IsValid(column int) bool {
	valid := filter.Choose(fc.Values, func(co ColumnOperator) bool {
		return co.Evaluate(column)
	})
	return len(valid.([]ColumnOperator)) > 0
}

// The UnmarshalText method allow FilterColumns to be used as an argument type
func (fc *FilterColumns) UnmarshalText(b []byte) error {
	re := regexp.MustCompile(`(<)?(\d+)(>)?`)
	s := string(b)
	values := strings.Split(s, ",")

	for _, val := range values {
		if !re.MatchString(val) {
			continue
		}

		// Parse expression
		matches := re.FindStringSubmatch(val)
		num, _ := strconv.Atoi(matches[2])

		if matches[1] == "" && matches[3] == "" {
			fc.Values = append(fc.Values, ColumnOperator{
				Column:     num,
				Comparison: Equal,
			})
		} else if matches[1] != "" {
			fc.Values = append(fc.Values, ColumnOperator{
				Column:     num,
				Comparison: LessThan,
			})
		} else if matches[3] != "" {
			fc.Values = append(fc.Values, ColumnOperator{
				Column:     num,
				Comparison: GreaterThan,
			})
		}
	}

	return nil
}

//The RenameColumns type stores which columns should be renamed
type RenameColumns struct {
	Values map[int]string
}

// The UnmarshalText method allow RenameColumns to be used as an argument type
func (rc *RenameColumns) UnmarshalText(b []byte) error {
	rc.Values = make(map[int]string)

	s := string(b)
	values := strings.Split(s, ",")

	for i := 0; i < len(values); i++ {
		columns := strings.Split(values[i], ":")
		if len(columns) < 2 {
			continue
		}

		idx, err := strconv.Atoi(columns[0])
		if err != nil {
			return fmt.Errorf("invalid index in %s", values[i])
		}

		rc.Values[idx] = columns[1]
	}

	return nil
}

// The args struct holds all argument types supported
var args struct {
	Columns   FilterColumns `arg:"-c" help:"Columns to show"`
	Ignore    FilterColumns `arg:"-i" help:"Columns to ignore"`
	Rename    RenameColumns `arg:"-r" help:"Columns to rename"`
	NoHeader  bool          `help:"Skip first line"`
	Filter    string        `help:"Filename containing the id to filter with"`
	FilterCol int           `help:"Column holding the value to filter for"`
	Inverse   bool          `help:"Inverts filter condition"`
}

// Returns a slice containing all values within the range going from 0 to size - 1
func newRange(size int) []int {
	col := 0
	cols := make([]int, size)
	filter.ApplyInPlace(cols, func(v int) int {
		x := v + col
		col++
		return x
	})
	return cols
}

// Returns a slice of strings without duplicated values
func unique(values []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range values {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// Finds a string in a slice
func find(needle string, haystack []string) bool {
	found := false
	for _, s := range haystack {
		if needle == s {
			found = true
			break
		}
	}

	return found
}

func main() {
	arg.MustParse(&args)

	// Check if a filter is provided
	filterValues := []string{}

	if args.Filter != "" {
		// Read filter values into a slice
		filterFilename := args.Filter
		filterFile, err := os.Open(filterFilename)

		if err != nil {
			log.Fatal(err)
		}

		filterReader := csv.NewReader(bufio.NewReader(filterFile))
		defer filterFile.Close()

		for {
			line, error := filterReader.Read()
			if error == io.EOF {
				break
			} else if error != nil {
				continue
			}

			filterValues = append(filterValues, line[0])
		}

		filterValues = unique(filterValues)
	}

	reader := csv.NewReader(os.Stdin)
	writer := csv.NewWriter(os.Stdout)

	var cols []int
	row := 0

	for {
		line, error := reader.Read()
		if error == io.EOF || error == errors.New("wrong number of fields") {
			break
		} else if error != nil {
			log.Fatal(error)
		}

		if row == 0 {
			// Build the column list
			cols = newRange(len(line))

			if len(args.Ignore.Values) > 0 {
				filter.DropInPlace(&cols, func(c int) bool {
					return args.Ignore.IsValid(c)
				})
			} else if len(args.Columns.Values) > 0 {
				filter.ChooseInPlace(&cols, func(c int) bool {
					return args.Columns.IsValid(c)
				})
			}

			// Skip first row
			if args.NoHeader {
				row++
				continue
			} else if len(args.Rename.Values) > 0 {
				// Rename if first line
				for idx, col := range args.Rename.Values {
					line[idx] = col
				}
			}
		}

		// Filter by id
		if args.Filter != "" && len(filterValues) > 0 {
			found := find(line[args.FilterCol], filterValues)

			// If the value is not found, read the next one
			if (!args.Inverse && !found) || (args.Inverse && found) {
				continue
			}
		}

		// Build line
		var out []string
		for _, column := range cols {
			out = append(out, line[column])
		}

		writer.Write(out)
		row++
	}

	writer.Flush()
}
