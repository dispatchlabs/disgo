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
package utils

import "time"

// ToNanoSeconds
func ToNanoSeconds(t time.Time) int64 {
	return t.UTC().UnixNano() / int64(time.Nanosecond)
}

// ToMicroSeconds
func ToMicroSeconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Microsecond)
}

// ToMilliSeconds
func ToMilliSeconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
