// generated by stringer -type=Status; DO NOT EDIT

package main

import "fmt"

const _Status_name = "StatusRedStatusSilverStatusGold"

var _Status_index = [...]uint8{0, 9, 21, 31}

func (i Status) String() string {
	if i < 0 || i >= Status(len(_Status_index)-1) {
		return fmt.Sprintf("Status(%d)", i)
	}
	return _Status_name[_Status_index[i]:_Status_index[i+1]]
}
