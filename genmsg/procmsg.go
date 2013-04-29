package main

import (
	"bufio"
	"fmt"
	"io"
//	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type parsedLine struct {
	is_constant  bool
	package_ string
	type_        string
	is_array     bool
	array_count  int
	name         string
	value        string
}

type msgParser struct {
	regex *regexp.Regexp

	data_type_package_index int
	data_type_index int
	is_array_index int
	array_count_index int
	data_name_index int
	const_type_package_index int
	const_type_index int
	const_name_index int
	const_value_index int
}

func newParser() *msgParser {
	parser := new(msgParser)

	line_start := "^"
	line_end := "$"
	word := "([[:alnum:]_]+)"
	type_ := "(" + word + "/)?" + word
	spc := "[[:space:]]+"
	optspc := "[[:space:]]*"
	opt_bracketed_number := "(\\[" + optspc + "([[:digit:]]*)" + optspc + "\\])?"
	opt_comment := "(#.*)?"
	constant := "([^#]*)"

	data_decl := type_ + opt_bracketed_number + spc + word
	const_decl := type_ + spc + word + optspc + "=" + optspc + constant
	data_or_const := "(" + data_decl + ")|(" + const_decl + ")"

	line_regex :=
		line_start + optspc + "(" + data_or_const + ")?" + optspc + opt_comment + line_end

	parser.data_type_package_index =  4
	parser.data_type_index =          5
	parser.is_array_index =           6
	parser.array_count_index =        7
	parser.data_name_index =          8
	parser.const_type_package_index = 11
	parser.const_type_index =         12
	parser.const_name_index =         13
	parser.const_value_index =        14

	var err error
	parser.regex, err = regexp.Compile(line_regex)
	if err != nil {
		fmt.Printf("shit, regex didn't work: %v\n", err)
		return nil
	}
	return parser
}

func (parser *msgParser) parseLine(line string) *parsedLine {
	line = strings.TrimSpace(line)

	match_result := parser.regex.FindStringSubmatch(line)
	//		for index, str := range match_result {
	//			fmt.Printf( "%v: %v\n", index, str )
	//		}
	if len(match_result) == 16 {
		result := new(parsedLine)
		if match_result[parser.data_name_index] != "" {
			result.name = match_result[ parser.data_name_index ]
			result.type_ = match_result[ parser.data_type_index ]
			result.package_ = match_result[ parser.data_type_package_index ]
			result.is_array = (match_result[ parser.is_array_index ] != "")
			array_count_string := match_result[ parser.array_count_index ]
			if( array_count_string != "" ) {
				count, err := strconv.Atoi( array_count_string )
				if( err == nil ) {
					result.array_count = count
				}
			}
		} else if match_result[ parser.const_name_index ] != "" {
			result.is_constant = true
			result.name = match_result[ parser.const_name_index ]
			result.type_ = match_result[ parser.const_type_index ]
			result.package_ = match_result[ parser.const_type_package_index ]
			result.value = match_result[ parser.const_value_index ]
		} else {
			return nil
		}
		result.rosToGo()
		return result
	}
	return nil
}

func camelCaseName( underscored_name string ) string {
	cap_next := true
	return strings.Map( func( char rune ) rune {
		if char == '_' {
			cap_next = true
			return -1
		}
		if cap_next {
			cap_next = false
			return unicode.ToUpper( char )
		}
		return char
	}, underscored_name )
}

func (line *parsedLine) rosToGo() {
	switch line.type_ {
	case "Header":
		line.package_ = "std_msgs"
	case "time":
		line.type_ = "Time"
		line.package_ = "ros"
	case "duration":
		line.type_ = "Duration"
		line.package_ = "ros"
	}
}

func processFile(package_ string, message_name string, input io.Reader, output io.Writer) {

	parser := newParser()

	buf_read := bufio.NewReader(input)

	imports := map[string] bool {}

	vars := make([]*parsedLine, 0)
	consts := make([]*parsedLine, 0)

	// Loop over input lines
	for {
		line, err := buf_read.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Printf("Error reading from stdin: %v\n", err)
			break
		}

		entry := parser.parseLine(line)
		if entry != nil {
			if entry.package_ != "" {
				imports[entry.package_] = true
			}
			if !entry.is_constant {
				vars = append( vars, entry )
			} else {
				consts = append( consts, entry )
			}
		}

		if err == io.EOF {
			break
		}
	}
	// Emit package line
	fmt.Fprintf(output, "package %v\n\n", package_)

	// Emit imports
	if len( imports ) > 0 {
		fmt.Fprintf(output, "import (\n")
		for import_, _ := range( imports ) {
			fmt.Fprintf(output, "\t\"%v\"\n", import_)
		}
		fmt.Fprintf(output, ")\n\n")
	}

	// Emit structure
	if len( vars ) > 0 {
		fmt.Fprintf(output, "type %v struct {\n", camelCaseName( message_name ))
		for _, var_ := range( vars ) {
			var package_, arrayness string
			if var_.package_ != "" {
				package_ = var_.package_ + "."
			}
			if var_.is_array {
				if var_.array_count > 0 {
					arrayness = fmt.Sprintf("[%v]", var_.array_count)
				} else {
					arrayness = "[]"
				}
			}
			go_type := package_ + var_.type_
			fmt.Fprintf(output,  "\t%v %v\n",
				camelCaseName( var_.name ),
				(arrayness + go_type) )
		}
		fmt.Fprintf(output, "}\n\n")
	}

	// Emit constants
	if len( consts ) > 0 {
		fmt.Fprintf(output, "const (\n")
		for _, const_ := range( consts ) {
			const_name := camelCaseName( message_name ) + "_" + camelCaseName( const_.name )
			if const_.package_ == "" {
				fmt.Fprintf(output, "\t%v %v = %v\n", const_name, const_.type_, const_.value)
			} else {
				fmt.Fprintf(output, "\t%v %v.%v = %v\n",
					const_name, const_.package_, const_.type_, const_.value)
			}
		}
		fmt.Fprintf(output, ")\n\n")
	}
}
