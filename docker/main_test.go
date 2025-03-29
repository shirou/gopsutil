// SPDX-License-Identifier: BSD-3-Clause
package docker

import (
	"fmt"
	"testing"
)

func TestSysAdvancedDockerInfo(_ *testing.T) {
	list, err := GetDockerIDList()
	if err != nil {
		fmt.Println(err)
	}
	for _, item := range list {
		fmt.Println(item)
	}
	/*docker,err := SysAdvancedDockerInfo()
	if err!= nil{
		fmt.Println(err)
	}
	fmt.Printf("%#v",docker)*/
}
