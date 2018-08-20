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
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/dispatchlabs/disgo/commons/types"
	"github.com/dispatchlabs/disgo/commons/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"golang.org/x/net/context"

	"github.com/processout/grpc-go-pool"
	"time"
	"github.com/pkg/errors"
	"github.com/patrickmn/go-cache"
)

var grpcServiceInstance *GrpcService
var grpcServiceOnce sync.Once

// GetGrpcService
func GetGrpcService() *GrpcService {
	grpcServiceOnce.Do(func() {
		opts := grpc.ServerOption(grpc.MaxRecvMsgSize(1024 * 1024 * 1024))
		grpcServiceInstance = &GrpcService{Port: int(types.GetConfig().GrpcEndpoint.Port), Server: grpc.NewServer(opts), running: false}
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
func (this *GrpcService) Go() {
	this.running = true
	listener, error := net.Listen("tcp", ":"+strconv.Itoa(this.Port))
	if error != nil {
		utils.Fatal(fmt.Sprintf("failed to listen: %v", error))
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


func GetGrpcConnection(address string, host string, port int64) (*grpc.ClientConn, error) {

	value, ok := GetCache().Get(fmt.Sprintf("dapos-grpc-pool-%s", address))
	// IF not found then setup one
	if !ok {
		setupConnectionPoolForPeer(address, host, port)
	}
	value, ok = GetCache().Get(fmt.Sprintf("dapos-grpc-pool-%s", address))
	if !ok {
		return nil, errors.New(fmt.Sprintf("unable to find GRPC pool for this delegate [address=%s]", address))
	}
	pool := value.(*grpcpool.Pool)
	clientConn, err := pool.Get(context.Background())
	if err != nil {
		utils.Error("Client connection error ", err)
	}
	defer clientConn.Close()
	return clientConn.ClientConn, nil

}

func setupConnectionPoolForPeer(address string, host string, port int64) {
	factory := func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())

		if err != nil {
			utils.Error("Failed to start gRPC connection: %v", err)
		}
		utils.Info("Connected to node at", host, port)
		return conn, err
	}
	pool, err := grpcpool.New(factory, 5, 5, time.Second* 5)
	if err != nil {
		utils.Error(err.Error())
	}
	GetCache().Set(fmt.Sprintf("dapos-grpc-pool-%s", address), pool, cache.NoExpiration)
}
