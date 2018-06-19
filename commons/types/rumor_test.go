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

func testMockRumor() *Rumor {
	return NewRumor("0f86ea981203b26b5b8244c8f661e30e5104555068a4bd168d3e3015db9bb25a", "3ed25f42484d517cdfc72cafb7ebc9e8baa52c2c", "9c242afd4f2dcaedcfb0cff2bb9c38b5811ed29c249f5b49f7759642a473d5fb")
}

// RomorVerify
func TestRumorVerify(t *testing.T) {
	rumor := testMockRumor()
	if rumor.Verify() {
		t.Log("rumor verified")
	} else {
		t.Error("cannot verify rumor")
	}
}
