package uuid

import "github.com/blend/go-sdk/exception"

// Error Classes
const (
	ErrParseInvalidUUIDInput = exception.Class("parse uuid: existing uuid is invalid")
	ErrParseEmpty            = exception.Class("parse uuid: input is empty")
	ErrParseInvalidLength    = exception.Class("parse uuid: input is an invalid length")
	ErrParseIllegalCharacter = exception.Class("parse uuid: illegal character")
)

// MustParse parses a uuid and will panic if there is an error.
func MustParse(corpus string) UUID {
	uuid := Empty()
	if err := ParseExisting(&uuid, corpus); err != nil {
		panic(err)
	}
	return uuid
}

// Parse parses a uuidv4 from a given string.
// valid forms are:
// - {xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx}
// - xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// - xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
func Parse(corpus string) (UUID, error) {
	uuid := Empty()
	if err := ParseExisting(&uuid, corpus); err != nil {
		return nil, err
	}
	return uuid, nil
}

// ParseExisting parses into an existing UUID.
func ParseExisting(uuid *UUID, corpus string) error {
	if len(corpus) == 0 {
		return exception.New(ErrParseEmpty)
	}
	if len(corpus)%2 == 1 {
		return exception.New(ErrParseInvalidLength)
	}
	if len(*uuid) != 16 {
		return exception.New(ErrParseInvalidUUIDInput)
	}
	var data = []byte(corpus)
	var c byte
	hex := [2]byte{}
	var hexChar byte
	var isHexChar bool
	var hexIndex, uuidIndex, di int

	for i := 0; i < len(data); i++ {
		c = data[i]
		if c == '{' && i == 0 {
			continue
		}
		if c == '{' {
			return exception.New(ErrParseIllegalCharacter).WithMessagef("at %d: %v", i, string(c))
		}
		if c == '}' && i != len(data)-1 {
			return exception.New(ErrParseIllegalCharacter).WithMessagef("at %d: %v", i, string(c))
		}
		if c == '}' {
			continue
		}

		if c == '-' && !(di == 8 || di == 12 || di == 16 || di == 20) {
			return exception.New(ErrParseIllegalCharacter).WithMessagef("at %d: %v", i, string(c))
		}
		if c == '-' {
			continue
		}

		hexChar, isHexChar = fromHexChar(c)
		if !isHexChar {
			return exception.New(ErrParseIllegalCharacter).WithMessagef("at %d: %v", i, string(c))
		}

		hex[hexIndex] = hexChar
		if hexIndex == 1 {
			(*uuid)[uuidIndex] = hex[0]<<4 | hex[1]
			uuidIndex++
			hexIndex = 0
		} else {
			hexIndex++
		}
		di++
	}
	if uuidIndex != 16 {
		return exception.New(ErrParseInvalidLength)
	}
	return nil
}

func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}

	return 0, false
}
