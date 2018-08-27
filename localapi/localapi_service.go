// /*
//  *    This file is part of DAPoS library.
//  *
//  *    The DAPoS library is free software: you can redistribute it and/or modify
//  *    it under the terms of the GNU General Public License as published by
//  *    the Free Software Foundation, either version 3 of the License, or
//  *    (at your option) any later version.
//  *
//  *    The DAPoS library is distributed in the hope that it will be useful,
//  *    but WITHOUT ANY WARRANTY; without even the implied warranty of
//  *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  *    GNU General Public License for more details.
//  *
//  *    You should have received a copy of the GNU General Public License
//  *    along with the DAPoS library.  If not, see <http://www.gnu.org/licenses/>.
//  */
package localapi

import (
	"sync"

	"github.com/dispatchlabs/disgo/commons/utils"
)

var localapiServiceInstance *LocalAPIService
var localapiServiceOnce sync.Once

type localAPIEvents struct {
	LocalAPIServiceInitFinished string
}

var (
	// Events - `dapos` events
	Events = localAPIEvents{
		LocalAPIServiceInitFinished: "LocalAPIServiceInitFinished",
	}
)

// GetLocalAPIService
func GetLocalAPIService() *LocalAPIService {
	localapiServiceOnce.Do(func() {
		localapiServiceInstance = &LocalAPIService{
			running: false,
		}
	})
	return localapiServiceInstance
}

// LocalAPIService -
type LocalAPIService struct {
	running bool
}

// IsRunning -
func (this *LocalAPIService) IsRunning() bool {
	return this.running
}

// Go -
func (this *LocalAPIService) Go() {
	this.running = true
	utils.Info("running, waiting for delegates sync")

	utils.Events().Raise(Events.LocalAPIServiceInitFinished)
}
