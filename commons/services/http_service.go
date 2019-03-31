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
			PublicApiEndpoint: *types.GetConfig().HttpEndpoint,
			PrivateApiPort:    types.GetConfig().LocalHttpApiPort,
			running:           false,
			router:            mux.NewRouter(),
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
	PublicApiEndpoint types.Endpoint
	PrivateApiPort    int
	running           bool
	router            *mux.Router
}

// IsRunning
func (this *HttpService) IsRunning() bool {
	return this.running
}

// Go -
func (this *HttpService) Go() {
	var wg sync.WaitGroup

	wg.Add(1)
	wg.Add(1)

	go func() {
		this.running = true
		listen := fmt.Sprintf("%s:%d", "", this.PublicApiEndpoint.Port) // FIX: commented "this.Endpoint.Host" for prod release
		utils.Info("listening on http://" + listen)
		cors := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true,
		})
		handler := cors.Handler(this.router)

		err := http.ListenAndServe(listen, utils.Gzip(handler))

		// QUESTION: this line is never reached
		utils.Error("unable to listen/serve HTTP [error=" + err.Error() + "]")

		wg.Done()
	}()

	go func() {
		listen := fmt.Sprintf("127.0.0.1:%d", this.PrivateApiPort)
		utils.Info("listening on http://" + listen)
		cors := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true,
		})
		handler := cors.Handler(this.router)

		err := http.ListenAndServe(listen, handler)

		// QUESTION: this line is never reached
		utils.Error("unable to listen/serve HTTP [error=" + err.Error() + "]")

		wg.Done()

	}()

	utils.Events().Raise(types.Events.HttpServiceInitFinished)

	wg.Wait()
}

// Error replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// This is an override of the default http.Error to set the header content type to application/json
func Error(responseWriter http.ResponseWriter, error string, code int) {
	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.Header().Set("X-Content-Type-Options", "nosniff")
	responseWriter.WriteHeader(code)
	fmt.Fprintln(responseWriter, error)
}
