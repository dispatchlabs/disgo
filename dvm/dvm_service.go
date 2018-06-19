/*
 *    This file is part of DVM library.
 *
 *    The DVM library is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    The DVM library is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with the DVM library.  If not, see <http://www.gnu.org/licenses/>.
 */
package dvm

import (
	"sync"

	"github.com/dispatchlabs/disgo/commons/utils"
)

var dvmServiceInstance *DVMService
var dvmServiceOnce sync.Once

type dvmEvents struct {
	DVMServiceInitFinished string
}

var (
	// Events - `dvm` events
	Events = dvmEvents{
		DVMServiceInitFinished: "DVMServiceInitFinished",
	}
)

// GetDVMService
func GetDVMService() *DVMService {
	dvmServiceOnce.Do(func() {
		dvmServiceInstance = &DVMService{running: false}
	})

	return dvmServiceInstance
}

// DVMService -
type DVMService struct {
	running bool
}

// IsRunning -
func (dvm *DVMService) IsRunning() bool {
	return dvm.running
}

// Go -
func (dvm *DVMService) Go(waitGroup *sync.WaitGroup) {
	dvm.running = true
	utils.Info("running")

	utils.Events().Raise(Events.DVMServiceInitFinished)
}
