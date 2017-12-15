/*
This file is part of the CHIP debugging bridge (CDB).
Copyright (C) 2018 Next Thing Co.

Gadget is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 2 of the License, or
(at your option) any later version.

Gadget is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Gadget.  If not, see <http://www.gnu.org/licenses/>.
*/

package libcdb


import (
	"os"
	"log"
	"net"
	"errors"
	"strings"
)


var (
	ip     = ""
	hostIp = ""
)


func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}


func EnsureIp() error {

	ifaces, _ := net.Interfaces()
	// handle err
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			var localIP net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				localIP = v.IP
				log.Printf("Found IP: %s", localIP.String())
				log.Printf("Searching for: %s", hostIp)
				if localIP.String() == hostIp {
					return nil
				}
			}
		}
	}

	return errors.New("Could not find Gadget IP")

}


func PrependToStrings(stringArray []string, prefix string) []string {

	if len(stringArray) == 0 || (len(stringArray) == 1 && stringArray[0] == "") {
		return []string{""}
	}

	for key, value := range stringArray {
		s := []string{prefix, value}
		stringArray[key] = strings.Join(s, "")
	}
	return stringArray
}

