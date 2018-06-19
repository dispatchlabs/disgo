/*
 *    This file is part of Disgo-Commons library.
 *
 *    The Disgo-Commons library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The Disgo-Commons library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the Disgo-Commons library.  If not, see <http://www.gnu.org/licenses/>.
 */
package utils

import (
	log "github.com/sirupsen/logrus"

	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func AddHeader(directory string, pkg string) {
	origHeaderFile := "LGPL-header.txt"
	header, _ := ReadLinesFromFile(origHeaderFile)

	fileList, _ := GetFilesFromDir(directory, ".go")

	replaced := make([]string, 0)
	for _, str := range header {
		result := strings.Replace(str, "%%==lib==%%", pkg, -1)
		replaced = append(replaced, result)
	}
	for _, file := range fileList {
		AddHeaderToFile(replaced, file)
	}
}

func GetFilesFromDir(searchDir string, fileExt string) ([]string, error) {
	fileList := make([]string, 0)
	e := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && filepath.Ext(path) == fileExt {
			fileList = append(fileList, path)
		}
		return err
	})

	if e != nil {
		panic(e)
	}

	for _, file := range fileList {
		fmt.Println(file)
	}

	return fileList, nil
}

func GetHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func ReadLinesFromFile(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	result := make([]string, 0, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return result, nil
}

// writeLines writes the lines to the given file.
func WriteLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}

func AddHeaderToFile(data []string, file string) {
	stringList := make([]string, 0, 0)

	for _, v := range data {
		stringList = append(stringList, v)
	}
	lines, err := ReadLinesFromFile(file)
	if err != nil {
		panic(err)
	}
	firstLine := lines[0]
	if CaseInsensitiveContains(firstLine, "copyright") {
		return
	}

	log.Info("Number of lines in second file = ", len(lines))
	if len(lines) > len(data) {
		matchedLines := checkForExisting(data, lines)

		if matchedLines == len(data) {
			log.Info("Identical header match.  Nothing to do")
			return
		} else if matchedLines != 0 && matchedLines != len(data) {
			log.Info("Header changed removing existing Header")
			lines = removeExistingHeader(lines)
		}
	}

	for _, v := range lines {
		stringList = append(stringList, v)
	}
	log.Info("Writing new file")
	WriteLines(stringList, file)
}

func CaseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

var HEADER_BEGIN = "/*"
var HEADER_END = "*/"

func removeExistingHeader(fileData []string) []string {
	beginIndex := 0
	endIndex := 0
	for i := 0; i < len(fileData); i++ {
		if fileData[i] == HEADER_BEGIN {
			beginIndex = i
		}
		if fileData[i] == HEADER_END {
			endIndex = i + 1
			break
		}
	}
	var stringList []string
	if beginIndex == 0 {
		stringList = fileData[endIndex:]
	} else {
		stringList = Join(fileData[0:beginIndex], fileData[endIndex:])
	}
	return stringList
}

func checkForExisting(header []string, fileData []string) int {
	matchCount := 0
	for i := 0; i < len(header); i++ {
		if header[i] == fileData[i] {
			matchCount++
		}
	}
	fmt.Printf("Match count = %d\n", matchCount)
	return matchCount
}

func Join(array1 []string, array2 []string) []string {
	newArray := make([]string, 0, 0)

	for _, v := range array1 {
		newArray = append(newArray, v)
	}
	newArray = append(newArray, "\n")

	for _, v := range array2 {
		newArray = append(newArray, v)
	}
	return newArray
}

func MergeFiles(file1 string, file2 string, outFile string) {
	lines, err := ReadLinesFromFile(file1)
	if err != nil {
		panic(err)
	}
	log.Info("Number of lines in file = ", len(lines))

	stringList := make([]string, 0, 0)

	for _, v := range lines {
		stringList = append(stringList, v)
	}
	stringList = append(stringList, "\n")

	lines, err = ReadLinesFromFile(file2)
	if err != nil {
		panic(err)
	}
	log.Info("Number of lines in second file = ", len(lines))

	for _, v := range lines {
		stringList = append(stringList, v)
	}
	WriteLines(stringList, outFile)
}
