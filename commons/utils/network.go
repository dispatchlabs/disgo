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

import (
	"net"
	"strings"
)

// GetLocalIP -
func GetLocalIP() string {
	return strings.Join(getLocalIPList(), ",")
}

func getLocalIPList() []string {
	var ipList = []string{}

	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			var ipAddress = ip.String()

			// var isUnspecified = ip.IsUnspecified()
			// var isLoopback = ip.IsLoopback()
			// var isMulticast = ip.IsMulticast()
			// var isInterfaceLocalMulticast = ip.IsInterfaceLocalMulticast()
			// var isLinkLocalMulticast = ip.IsLinkLocalMulticast()
			// var isLinkLocalUnicast = ip.IsLinkLocalUnicast()
			// var isGlobalUnicast = ip.IsGlobalUnicast()

			if ip.IsGlobalUnicast() {
				ipList = append(ipList, ipAddress)
			}
		}
	}

	return ipList

	// name, err := os.Hostname()
	// if err != nil {
	// 	fmt.Printf("Oops: %v\n", err)
	// 	return ""
	// }

	// addrs, err := net.LookupHost(name)
	// if err != nil {
	// 	fmt.Printf("Oops: %v\n", err)
	// 	return ""
	// }
	// fmt.Printf("Local IP: %s\n", addrs[0])

	// return addrs[0]
}
