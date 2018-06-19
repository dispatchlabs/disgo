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
package services

import (
	"net/http"
	"sync"

	"fmt"

	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var httpServiceInstance *HttpService
var httpServiceOnce sync.Once

// GetHttpService
func GetHttpService() *HttpService {
	httpServiceOnce.Do(func() {
		httpServiceInstance = &HttpService{
			Endpoint: *types.GetConfig().HttpEndpoint,
			running:  false,
			router:   mux.NewRouter(),
		}
	})
	return httpServiceInstance
}

// GetHttpRouter
func GetHttpRouter() *mux.Router {
	return GetHttpService().router
}

// HttpService
type HttpService struct {
	Endpoint types.Endpoint
	running  bool
	router   *mux.Router
}

// IsRunning
func (this *HttpService) IsRunning() bool {
	return this.running
}

// Go -
func (this *HttpService) Go(waitGroup *sync.WaitGroup) {
	this.running = true
	listen := fmt.Sprintf("%s:%d", this.Endpoint.Host, this.Endpoint.Port)
	utils.Info("listening on http://" + listen)
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})
	handler := cors.Handler(this.router)

	utils.Events().Raise(Events.HttpServiceInitFinished)

	error := http.ListenAndServe(listen, handler)

	// QUESTION: this line is never reached
	utils.Error("unable to listen/serve HTTP [error=" + error.Error() + "]")
}
