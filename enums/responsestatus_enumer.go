// Code generated by "enumer -type=ResponseStatus -json enums"; DO NOT EDIT.

package enums

import (
	"encoding/json"
	"fmt"
)

const _ResponseStatusName = "SuccessError"

var _ResponseStatusIndex = [...]uint8{0, 7, 12}

func (i ResponseStatus) String() string {
	if i < 0 || i >= ResponseStatus(len(_ResponseStatusIndex)-1) {
		return fmt.Sprintf("ResponseStatus(%d)", i)
	}
	return _ResponseStatusName[_ResponseStatusIndex[i]:_ResponseStatusIndex[i+1]]
}

var _ResponseStatusValues = []ResponseStatus{0, 1}

var _ResponseStatusNameToValueMap = map[string]ResponseStatus{
	_ResponseStatusName[0:7]:  0,
	_ResponseStatusName[7:12]: 1,
}

// ResponseStatusString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func ResponseStatusString(s string) (ResponseStatus, error) {
	if val, ok := _ResponseStatusNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to ResponseStatus values", s)
}

// ResponseStatusValues returns all values of the enum
func ResponseStatusValues() []ResponseStatus {
	return _ResponseStatusValues
}

// IsAResponseStatus returns "true" if the value is listed in the enum definition. "false" otherwise
func (i ResponseStatus) IsAResponseStatus() bool {
	for _, v := range _ResponseStatusValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for ResponseStatus
func (i ResponseStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for ResponseStatus
func (i *ResponseStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("ResponseStatus should be a string, got %s", data)
	}

	var err error
	*i, err = ResponseStatusString(s)
	return err
}
