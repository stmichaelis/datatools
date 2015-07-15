/*  csv2vw converts data formated as comma separated values to the vowpal wabbit input format.

    Copyright (C) 2015 Stefan Michaelis <info@stefan-michaelis.name>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as
    published by the Free Software Foundation, either version 3 of the
    License, or (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type intmap map[int]bool

var (
	ofile, ifile string
	labelS       string
	nominalS     string
	removeS      string
	idS          string
	importanceS  string
	label        int
	nominal      intmap
	remove       intmap
	id           int
	importance   int
	NCPU         int
	hasheader    bool
	useheader    bool
	headerrow    []string
	header       map[string]int
	predictonly  bool
	quiet        bool
)

func init() {
	nominal = make(map[int]bool)
	remove = make(map[int]bool)
	header = make(map[string]int)

	// Set base dir via commandline parameter
	flag.StringVar(&ofile, "o", "", "-o <Outputfile>")
	flag.StringVar(&ifile, "i", "", "-i <Inputfile>")

	flag.StringVar(&labelS, "l", "0", "Column index (starting with 0) or column name (needs header=true) of label attribute. Set to -1 to use last value per row as label. For numerical column names prepend with _. Example: -l _23 for using column named 23 as label column instead of column with index 23.")
	flag.StringVar(&idS, "id", "-1", " Column index (starting with 0) or column name (needs header=true) of id (tag) attribute. For numerical column names prepend with _. Set to -1 for data without an index attribute.")
	flag.StringVar(&importanceS, "w", "-1", " Column index (starting with 0) or column name (needs header=true) of importance weight column. For numerical column names prepend with _. Set to -1 for data without a importance column.")
	flag.StringVar(&nominalS, "n", "", "Column indices (starting with 0) or column names (needs header=true) of nominal/categorial attributes. List separated by comma. Example: -n 1,2,5,6. For numerical column names prepend with _.")
	flag.StringVar(&removeS, "r", "", "Remove comlumns with given indices (starting with 0) or column names (needs header=true) . List separated by comma. Example: -r 2,5,8,9. For numerical column names prepend with _.")
	flag.IntVar(&NCPU, "ncpu", 0, "Set number of cores to use. Values <= 0 use all availabe cores. Important: Set to 1 for preserving the same example order as in the input data set.")
	flag.BoolVar(&hasheader, "header", true, "Use first line in CSV file as header.")
	flag.BoolVar(&useheader, "headernames", false, "Use names of header columns as attribute names instead of index number. May increase output file size. Implies -header=true.")
	flag.BoolVar(&predictonly, "nolabel", false, "Omit label in output data for performing prediction only in VW.")
	flag.BoolVar(&quiet, "q", false, "Quiet mode.")
}

// TODO:
// Automatic parsing of Quotes around nominal attributes, i.e. via double quotes
// Specify separator
// Daemon mode for streaming data to VW

func transformLine(lines chan []string, newlines chan string, wg *sync.WaitGroup) {
	for line := range lines {
		var nl bytes.Buffer
		if !predictonly {
			if labelS == "-1" {
				label = len(line) - 1
			}
			nl.WriteString(line[label])
			nl.WriteString(" ")
		}

		if importance > -1 {
			nl.WriteString(line[importance])
			nl.WriteString(" ")
		}

		if id > -1 {
			nl.WriteString(line[id])
		}
		// Label separator and name space
		nl.WriteString("|n ")
		for i, v := range line {
			// Sparse attributes
			if v == "" {
				continue
			}

			_, r := remove[i]
			if i != label && i != id && i != importance && !r {
				_, n := nominal[i]
				// Sparse attributes for 0 values of non-nominal attributes
				if !n {
					f, err := strconv.ParseFloat(v, 64)
					if err != nil || f == 0 {
						continue
					}
				}
				if !useheader {
					// Write attribute index
					nl.WriteString(strconv.Itoa(i))
				} else {
					// Write attribute (column) name
					nl.WriteString(headerrow[i])
				}

				if !n {
					// Numerical, non-nominal attribute
					nl.WriteString(":")
				} else {
					// Nominal attribute
					nl.WriteString("=")
				}
				nl.WriteString(v)
				nl.WriteString(" ")
			}
		}

		nl.WriteString("\n")
		newlines <- nl.String()
	}
	wg.Done()
}

func parseIndex(s string) int {
	tmp, err := strconv.Atoi(s)
	if err != nil {
		// Not parseable integer, try to find via column name
		// Remove leading _, if any
		s = strings.TrimPrefix(s, "_")
		var ok bool
		tmp, ok = header[s]
		if !ok {
			se := "Error parsing %v as column index."
			if hasheader {
				se += " And did not find a matching column name from header."
			}
			log.Fatalf(se, s)

		}
	}
	return tmp
}

func (i intmap) parseIndices(s string) {
	for _, v := range strings.Split(s, ",") {
		if v != "" {
			i[parseIndex(v)] = true
		}
	}
}

func main() {
	flag.Parse()
	if useheader {
		hasheader = true
	}
	if quiet {
		log.SetOutput(ioutil.Discard)
	}

	log.Printf("Started data conversion.\n")

	var err error
	// Use all available cores
	if NCPU <= 0 {
		NCPU = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(NCPU)

	// Open input/output files
	in, err := os.Open(ifile)
	if err != nil {
		log.Fatal(err)
	}
	csvin := csv.NewReader(in)
	// Allow rows with variable number of fields
	csvin.FieldsPerRecord = -1
	out, err := os.Create(ofile)
	if err != nil {
		log.Fatal(err)
	}
	vwout := bufio.NewWriter(out)

	defer func() {
		in.Close()
		out.Close()
	}()

	var countread int64

	// Read and parse first line
	if hasheader {
		headerrow, err = csvin.Read()
		countread++
		if err != nil {
			log.Fatal(err)
		}
		for i, v := range headerrow {
			header[v] = i
		}
	}

	// Parsing indices or header name parameters
	label = parseIndex(labelS)
	id = parseIndex(idS)
	importance = parseIndex(importanceS)
	nominal.parseIndices(nominalS)
	remove.parseIndices(removeS)

	// Start one go routine per core
	lines := make(chan []string, NCPU*10)
	newlines := make(chan string, NCPU*10)
	wg := new(sync.WaitGroup)
	for i := 0; i < NCPU; i++ {
		wg.Add(1)
		go transformLine(lines, newlines, wg)
	}
	// Wait for all go routines to finish, then close newlines result channel
	go func() {
		wg.Wait()
		close(newlines)
	}()

	go func() {
		// *** Iterate over all lines ***
		var line []string
		for {
			line, err = csvin.Read()
			if err != nil {
				break
			}
			// Send current line to go routines
			lines <- line
			// Count all lines in input file
			countread++
			if countread%10000 == 0 && !quiet {
				fmt.Println(countread)
			}
		}
		close(lines)
		if err != io.EOF {
			log.Fatal(err)
		}
	}()

	// Write all processed lines
	var countwritten int64
	for l := range newlines {
		vwout.WriteString(l)
		countwritten++
	}

	vwout.Flush()

	log.Printf("Done converting with reading %v and writing %v lines.\n", countread, countwritten)
}
