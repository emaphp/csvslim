# csvslim

A small utility for manipulating CSV files.

## About

`csvslim` is a small program that allows you to manipulate CSV files. It includes the following features:

 * Filter columns out.
 * Renaming columns.
 * Filtering rows.

## Build

```
 go get "github.com/alexflint/go-arg"
 go get "robpike.io/filter"
 go build csvslim.go
```

## Usage

```
csvslim [FLAGS] < input.csv
```

`csvslim` supports the following flags:

 * `-c COLUMN1,COLUMN2,...`: Shows only the entered columns.
 * `-i COLUMN1,COLUMN2,...`: Removes entered columns from the output.
 * `-r "COLUMN1:NAME1,COLUMN2:NAME2,..."`: Renames columns on the first line.
 * `--noheader`: Removes the first line from the output.
 * `--filter FILE`: Filters the input using the values within `FILE`. By default, values on the input file are compared with the ones listed on the first column.
 * `--filtercol COLUMN`: Specify the column number holding the values that should be compared against the ones on the filter file.
 * `--inverse`: Inverts filtering condition.
 
Both `-c` and `-i` support including a comparison operator for removing columns within an interval.

## Examples

### Show only the second and fourth columns

```
csvslim -c 1,3 < input.csv
```

### Show only the first 3 columns, and fifth

```
csvslim -c "<3,4" < input.csv
```

### Ignore the first column and anything past the fifth

```
csvslim -c "<1,4>" < input.csv
```

### Rename first column to "Code" and third to "Retail Price"

```
csvslim -r "0:Code,2:Retail Price" < input.csv
```

### Show only values matching the ones on the filter file

```
csvslim --filter filter.csv < input.csv
```

### Show only values not matching the ones on the filter file

```
csvslim --filter filter.csv --inverse < input.csv
```

### Show only values matching the filter file but compare against second column

```
csvslim --filter filter.csv --filtercol 1 < input.csv
```

## License

Licensed under MIT License.
