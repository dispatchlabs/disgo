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
	"net"
	"strconv"
	"sync"

	"github.com/dispatchlabs/commons/types"
	"github.com/dispatchlabs/commons/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var grpcServiceInstance *GrpcService
var grpcServiceOnce sync.Once

// GetGrpcService
func GetGrpcService() *GrpcService {
	grpcServiceOnce.Do(func() {
		grpcServiceInstance = &GrpcService{Port: int(types.GetConfig().GrpcEndpoint.Port), Server: grpc.NewServer(), running: false}
	})
	return grpcServiceInstance
}

// GrpcService
type GrpcService struct {
	Port    int
	Server  *grpc.Server
	running bool
}

// IsRunning
func (this *GrpcService) IsRunning() bool {
	return this.running
}

// Go
func (this *GrpcService) Go(waitGroup *sync.WaitGroup) {
	this.running = true
	listener, error := net.Listen("tcp", ":"+strconv.Itoa(this.Port))
	if error != nil {
		utils.Fatal("failed to listen: %v", error)
		this.running = false
		return
	}

	// Serve.
	utils.Info("listening on " + strconv.Itoa(this.Port))
	reflection.Register(this.Server)

	utils.Events().Raise(Events.GrpcServiceInitFinished)

	if error := this.Server.Serve(listener); error != nil {
		utils.Fatal("failed to serve: %v", error)
		this.running = false
	}
}
