package dgowrapper

import "strconv"

// IsNumber makes sure string is number
func IsNumber(v interface{}) error {
	_, err := strconv.Atoi(v.(string))
	if err != nil {
		return ErrFilterFailed
	}

	return nil
}
