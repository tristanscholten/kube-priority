package annotation

import (
	"fmt"
	"strconv"
	"strings"
)

type Bounds struct {
	Minimum          int64
	MaximumExclusive int64
}

type Parsed struct {
	Raw   string
	Value int32
}

func Parse(raw string, b Bounds) (Parsed, error) {
	if raw == "" {
		return Parsed{}, fmt.Errorf("annotation value must not be empty")
	}
	if strings.TrimSpace(raw) != raw {
		return Parsed{}, fmt.Errorf("annotation value must not contain leading or trailing whitespace")
	}
	for _, r := range raw {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return Parsed{}, fmt.Errorf("annotation value must not contain whitespace")
		}
	}
	if strings.HasPrefix(raw, "+") {
		return Parsed{}, fmt.Errorf("annotation value must be a base-10 integer without a plus sign")
	}
	v64, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return Parsed{}, fmt.Errorf("annotation value must be a signed 32-bit base-10 integer: %w", err)
	}
	if v64 < b.Minimum {
		return Parsed{}, fmt.Errorf("priority value %d is below configured minimum %d", v64, b.Minimum)
	}
	if v64 >= b.MaximumExclusive {
		return Parsed{}, fmt.Errorf("priority value %d must be less than %d", v64, b.MaximumExclusive)
	}
	return Parsed{Raw: raw, Value: int32(v64)}, nil
}
