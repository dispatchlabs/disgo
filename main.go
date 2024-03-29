/*
 *    This file is part of Disgo library.
 *
 *    The Disgo library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The Disgo library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the Disgo library.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"github.com/dispatchlabs/disgo/bootstrap"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/dispatchlabs/disgo/commons/types"
)

var version string
var date string

func main() {
	utils.InitMainPackagePath()
	utils.InitializeLogger()
	if version != "" {
		types.SetVersion(version, date)
	} 

	server := bootstrap.NewServer()
	server.Go()
}
