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

func main() {

	privateKey := "2ce5279c21e080250d152054d448e90d2952fec5fd0bccaa1d7c9886d40b45cc"
	from := "dbae0d9e9b819c41ab7801a748f9c928fc9cf317"


	//privateKey string, from, to string, value string, hertz int64, timeInMiliseconds int64

	t, _ := types.NewTransferTokensTransaction(privateKey, from, "3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c", "10000000", 0, 0)
	s := t.String()

	utils.Info(s)
	utils.InitMainPackagePath()
	utils.InitializeLogger()
	server := bootstrap.NewServer()
	server.Go()
}
