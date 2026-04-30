package utils

import (
	"strings"
	"time"
)

const DateFormat = "2006-01-02 15:04"

type CustomDate time.Time

func (c *CustomDate) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), "\"")

	if str == "null" || str == "" {
		return nil
	}

	parsedTime, err := time.Parse(DateFormat, str)
	if err != nil {
		return err
	}

	*c = CustomDate(parsedTime)
	return nil
}

func (c *CustomDate) Time() time.Time {
	if c == nil {
		return time.Time{}
	}
	return time.Time(*c)
}
