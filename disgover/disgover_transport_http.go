/*	
 *    This file is part of Disgover library.
 *
 *    The Disgover library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The Disgover library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the Disgover library.  If not, see <http://www.gnu.org/licenses/>.
*/
package disgover

import (
	"fmt"
	"github.com/dispatchlabs/commons/services"
	"io/ioutil"
	"net/http"
)

func (this *DisGoverService) WithHttp() *DisGoverService {
	services.GetHttpRouter().HandleFunc("/v1/ping", this.pingPongHandler).Methods("POST")
	return this
}

func (this *DisGoverService) pingPongHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, _ := ioutil.ReadAll(request.Body)

	fmt.Println(string(body))

	responseWriter.Write([]byte(fmt.Sprintf(
		"PONG-From: %s @ %s:%d",
		this.ThisNode.Address,
		this.ThisNode.Endpoint.Host,
		this.ThisNode.Endpoint.Port,
	)))
}
