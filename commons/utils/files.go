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
	"os"
)

// Exists -
func Exists(name string) bool {
	_, error := os.Stat(name)
	return !os.IsNotExist(error)
}

// GetConfigDir -
func GetConfigDir() string {
	directoryName := "." + string(os.PathSeparator) + "config"
	if !Exists(directoryName) {
		err := os.MkdirAll(directoryName, 0755)
		if err != nil {
			Error(fmt.Sprintf("unable to create directory %s", directoryName), err)
			panic(err)
		}
	}
	return directoryName
}

/*
// user, error := user.Current()
currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
if err != nil {
	log.WithFields(log.Fields{
		"method": GetCallingFuncName() + fmt.Sprintf(" -> %s", err),
	}).Fatal("unable to get current directory")

	panic(err)
}

// return user.HomeDir + string(os.PathSeparator) + ".disgo"

var configFolder = currentDir + string(os.PathSeparator) + "config"
os.MkdirAll(configFolder, os.ModePerm)

return configFolder
*/
