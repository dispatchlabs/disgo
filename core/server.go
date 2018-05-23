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
package core

import (
	"sync"

	"github.com/dispatchlabs/dvm"

	"github.com/dispatchlabs/commons/services"
	"github.com/dispatchlabs/commons/types"
	"github.com/dispatchlabs/commons/utils"
	"github.com/dispatchlabs/dapos"
	"github.com/dispatchlabs/disgover"
)

const (
	Version = "1.0.0"
)

// Server -
type Server struct {
	services []types.IService
}

// NewServer -
func NewServer() *Server {
	utils.InitializeLogger()

	// Load Keys
	if _, _, err := loadKeys(); err != nil {
		utils.Error("unable to keys: " + err.Error())
	}

	return &Server{}
}

// Go
func (server *Server) Go() {
	utils.Info("booting Disgo v" + Version + "...")

	// Add services.
	server.services = append(server.services, dvm.GetDVMService())
	server.services = append(server.services, services.GetDbService())
	server.services = append(server.services, disgover.GetDisGoverService().WithGrpc().WithHttp())
	server.services = append(server.services, dapos.GetDAPoSService().WithGrpc().WithHttp())
	server.services = append(server.services, services.GetHttpService())
	server.services = append(server.services, services.GetGrpcService())

	// Run services.
	var waitGroup sync.WaitGroup
	for _, service := range server.services {
		utils.Info("starting " + utils.GetStructName(service) + "...")
		go service.Go(&waitGroup)
		waitGroup.Add(1)
	}
	waitGroup.Wait()
}
