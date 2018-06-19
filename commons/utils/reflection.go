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
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
)

var strToReplace = "github.com/dispatchlabs"
var strToReplaceWith = "main"

var mainPackagePath string

// InitMainPackagePath - to be called in `main()`
func InitMainPackagePath() {
	_, filename, _, ok := runtime.Caller(1)
	if ok {
		mainPackagePath = path.Dir(filename) + "/"
		fmt.Println(mainPackagePath)

		mainPackagePath = strings.Replace(mainPackagePath, strToReplace, strToReplaceWith, 1)
	}
}

// GetPathForThisPackage -
func GetPathForThisPackage() string {
	_, filename, _, ok := runtime.Caller(1)
	if ok {
		return path.Dir(filename) + "/"
	}

	return "./"
}

// GetStructName - returns struct name
func GetStructName(i interface{}) string {
	var t = reflect.TypeOf(i)

	var result string

	if t.Kind() == reflect.Ptr {
		result = t.Elem().Name()
	} else {
		result = t.Name()
	}

	return strings.Replace(result, strToReplace, strToReplaceWith, 1)
}

// GetPackageName - returns package name
func GetPackageName(i interface{}) string {
	var t = reflect.TypeOf(i)

	var result string

	if t.Kind() == reflect.Ptr {
		result = t.Elem().PkgPath()
	} else {
		result = t.PkgPath()
	}

	return strings.Replace(result, strToReplace, strToReplaceWith, 1)
}

// GetPackageNameWithStruct - returns package name with struct
func GetPackageNameWithStruct(i interface{}) string {
	var t = reflect.TypeOf(i)

	var result string

	if t.Kind() == reflect.Ptr {
		result = t.Elem().PkgPath() + "/" + t.Elem().Name()
	} else {
		result = t.PkgPath() + "/" + t.Name()
	}

	return strings.Replace(result, strToReplace, strToReplaceWith, 1)
}

// GetCallingFuncName - returns package + function name at runtime
func GetCallingFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	var functionName = runtime.FuncForPC(pc).Name()

	functionName = strings.Replace(functionName, strToReplace, strToReplaceWith, 1)
	functionName = strings.Replace(functionName, "(", "", 1)
	functionName = strings.Replace(functionName, ")", "", 1)
	functionName = strings.Replace(functionName, "*", "", 1)

	return functionName + "()"
}

// GetFuncName
func GetFuncName(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	var functionName = runtime.FuncForPC(pc).Name()

	functionName = strings.Replace(functionName, strToReplace, strToReplaceWith, 1)
	functionName = strings.Replace(functionName, "(", "", 1)
	functionName = strings.Replace(functionName, ")", "", 1)
	functionName = strings.Replace(functionName, "*", "", 1)

	return functionName + "()"
}

// GetCallStackWithFileAndLineNumber - traces a call with line number
func GetCallStackWithFileAndLineNumber() string {
	mainPackagePath = "/home/nicu/go/src/github.com/dispatchlabs"

	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])

	var logLine = ""

	frame, more := frames.Next()
	for more {
		frame.File = strings.Replace(frame.File, mainPackagePath, "", 1)
		// logLine += fmt.Sprintf("%s,:%d %s", frame.File, frame.Line, frame.Function)
		logLine += fmt.Sprintf("%s | ", frame.Function)

		frame, more = frames.Next()
	}

	return logLine
}
