package jra

import (
	"fmt"
	"strconv"
	"strings"
)

type commaSeparatedInt int

func (csn *commaSeparatedInt) UnmarshalXPath(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	s := strings.ReplaceAll(string(text), ",", "")
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid number string found. text=%q", string(text))
	}
	*csn = commaSeparatedInt(n)
	return nil
}
