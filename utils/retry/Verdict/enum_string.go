// Code generated by "stringer -type=Enum"; DO NOT EDIT.

package Verdict

import "strconv"

const _Enum_name = "DoneRetryAbort"

var _Enum_index = [...]uint8{0, 4, 9, 14}

func (i Enum) String() string {
	if i < 0 || i >= Enum(len(_Enum_index)-1) {
		return "Enum(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Enum_name[_Enum_index[i]:_Enum_index[i+1]]
}
