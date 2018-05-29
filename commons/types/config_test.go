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
package types

import "testing"

func TestGetConfig(t *testing.T) {

	var c *Config

	c = GetConfig()

	if c.HttpEndpoint == nil {
		t.Error ("Config Host IP is nil")
	}
	if c.HttpEndpoint.Port != 1975 {
		t.Error( "Disgo Config PORT NOT defaulted to 1975 - we better have a good reason to change this" )
	}
	if c.SeedEndpoints == nil {
		t.Error ("Seed list just can't be nil.  That would be just silly.  I mean, where ya goona go ?")
	}
}

