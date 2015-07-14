# CSV2VW
Utility to convert from CSV data files to VW Vowpal Wabbit format. 

Use `go get github.com/stmichaelis/datatools` to download and build the conversion tool. Run `csv2vw -h` from the GOPATH bin directory to see all comandline options.

## Example
`csv2vw -i input.csv -o output.csv -l 2 -i ID_ATTRIBUTE -n 4,5,6 -r 8`

Convert `input.csv` to VW format using column with index 2 as training attribute, column named `ID_ATTRIBUTE` as tag, treat columns 4,5,6 as nominal values and remove column 8 from output file.

## Comandline options

  ```
  -i="": -i <Inputfile>
  ```
  ```
  -o="": -o <Outputfile>
  ```
```
  -header=true: Use first line in CSV file as header.
```
  ```
  -l="0": 
  Column index (starting with 0) or column name (needs header=true) 
  of label attribute. Set to -1 to use last value per row as label. 
  For numerical column names prepend with _. 
  Example: -l _23 for using column named 23 as label column instead of column with index 23.
  ```
```
  -id="-1":  
  Column index (starting with 0) or column name (needs header=true) 
  of id (tag) attribute. For numerical column names prepend with _. 
  Set to -1 for data without an index attribute.
  ```

  ```
  -n="": 
  Column indices (starting with 0) or column names (needs header=true) 
  of nominal/categorial attributes. List separated by comma. 
  Example: -n 1,2,5,6. For numerical column names prepend with _.
  ```
  ```
  -r="": 
  Remove comlumns with given indices (starting with 0) or 
  column names (needs header=true) . List separated by comma. 
  Example: -r 2,5,8,9. For numerical column names prepend with _.
  ```
  ```
  -w="-1":  
  Column index (starting with 0) or column name (needs header=true) 
  of importance weight column. For numerical column names prepend with _. 
  Set to -1 for data without a importance column.
```
  ```
  -ncpu=0: 
  Set number of cores to use. Values <= 0 use all availabe cores. 
  Important: Set to 1 for preserving the same example order as in the input data set.
  ```
  ```
  -nolabel=false: Omit label in output data for performing prediction only in VW.
  ```
  ```
  -headernames=false: 
  Use names of header columns as attribute names 
  instead of index number. May increase output file size. 
  Implies -header=true.
  ```
  The headernames flag may be helpful in case you want to process csv files with a different number and ordering of columns, e.g. train and test files where the column indices differ due to a missing label.
  ```
  -q=false: Quiet mode.
  ```

## LICENSE

Copyright (C) 2015 Stefan Michaelis <http://www.stefan-michaelis.name>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.
