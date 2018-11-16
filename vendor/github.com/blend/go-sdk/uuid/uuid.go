package uuid

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/blend/go-sdk/exception"
)

var (
	byteGroups = []int{8, 4, 4, 4, 12}

	byteGroupSeparatorOffsets = []int{8, 12, 16, 20}

	hextable = [16]byte{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'a', 'b', 'c', 'd', 'e', 'f',
	}
)

// ErrInvalidScanSource is an error returned by scan.
const (
	ErrInvalidScanSource exception.Class = "uuid: invalid scan source"
)

// Empty returns an empty uuid block.
func Empty() UUID {
	return UUID(make([]byte, 16))
}

// UUID represents a unique identifier conforming to the RFC 4122 standard.
// UUIDs are a fixed 128bit (16 byte) binary blob.
type UUID []byte

// Equal returns if a uuid is equal to another uuid.
func (uuid UUID) Equal(other UUID) bool {
	return bytes.Equal(uuid[0:], other[0:])
}

// Compare returns a comparison between the two uuids.
func (uuid UUID) Compare(other UUID) int {
	return bytes.Compare(uuid[0:], other[0:])
}

// ToFullString returns a "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" hex representation of a uuid.
func (uuid UUID) ToFullString() string {
	if len(uuid) == 0 {
		return ""
	}
	b := []byte(uuid)
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		b[:4], b[4:6], b[6:8], b[8:10], b[10:],
	)
}

// ToShortString returns a hex representation of the uuid.
func (uuid UUID) ToShortString() string {
	return hex.EncodeToString([]byte(uuid))
}

// String is an alias for `ToShortString`.
func (uuid UUID) String() string {
	return hex.EncodeToString([]byte(uuid))
}

// Version returns the version byte of a uuid.
func (uuid UUID) Version() byte {
	return uuid[6] >> 4
}

// Format allows for conditional expansion in printf statements
// based on the token and flags used.
func (uuid UUID) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, uuid.ToFullString())
			return
		}
		io.WriteString(s, uuid.String())
	case 's':
		io.WriteString(s, uuid.String())
	case 'q':
		fmt.Fprintf(s, "%b", uuid.Version())
	}
}

// IsZero returns if the uuid is unset.
func (uuid UUID) IsZero() bool {
	if len(uuid) == 0 {
		return true
	}
	return bytes.Equal([]byte(uuid), []byte(Empty()))
}

// IsV4 returns true iff uuid has version number 4, variant number 2, and length 16 bytes
func (uuid UUID) IsV4() bool {
	if len(uuid) != 16 {
		return false
	}
	// check that version number is 4
	if (uuid[6]&0xf0)^0x40 != 0 {
		return false
	}
	// check that variant is 2
	return (uuid[8]&0xc0)^0x80 == 0
}

// Marshal implements bytes marshal.
func (uuid UUID) Marshal() ([]byte, error) {
	if len(uuid) == 0 {
		return nil, nil
	}
	return []byte(uuid), nil
}

// MarshalTo marshals the uuid to a buffer.
func (uuid UUID) MarshalTo(data []byte) (n int, err error) {
	if len(uuid) == 0 {
		return 0, nil
	}
	copy(data, uuid)
	return 16, nil
}

// Unmarshal implements bytes unmarshal.
func (uuid *UUID) Unmarshal(data []byte) error {
	if len(data) == 0 {
		uuid = nil
		return nil
	}
	id := UUID(make([]byte, 16))
	copy(id, data)
	*uuid = id
	return nil
}

// Size returns the size of the uuid.
func (uuid *UUID) Size() int {
	if uuid == nil {
		return 0
	}
	if len(*uuid) == 0 {
		return 0
	}
	return 16
}

// MarshalJSON marshals a uuid as json.
func (uuid UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString([]byte(uuid)))
}

// UnmarshalJSON unmarshals a uuid from json.
func (uuid *UUID) UnmarshalJSON(corpus []byte) error {
	if len(*uuid) == 0 {
		(*uuid) = Empty()
	}
	raw := strings.TrimSpace(string(corpus))
	raw = strings.TrimPrefix(raw, "\"")
	raw = strings.TrimSuffix(raw, "\"")
	return ParseExisting(uuid, raw)
}

// MarshalYAML marshals a uuid as yaml.
func (uuid UUID) MarshalYAML() (interface{}, error) {
	return uuid.ToFullString(), nil
}

// UnmarshalYAML unmarshals a uuid from yaml.
func (uuid *UUID) UnmarshalYAML(unmarshaler func(interface{}) error) error {
	if len(*uuid) == 0 {
		(*uuid) = Empty()
	}

	var corpus string
	if err := unmarshaler(&corpus); err != nil {
		return err
	}

	raw := strings.TrimSpace(string(corpus))
	raw = strings.TrimPrefix(raw, "\"")
	raw = strings.TrimSuffix(raw, "\"")
	return ParseExisting(uuid, raw)
}

// Scan scans a uuid from a db value.
func (uuid *UUID) Scan(src interface{}) error {
	if len(*uuid) == 0 {
		(*uuid) = Empty()
	}
	switch src.(type) {
	case string:
		return ParseExisting(uuid, src.(string))
	case []byte:
		return ParseExisting(uuid, string(src.([]byte)))
	}
	return exception.New(ErrInvalidScanSource).WithMessagef("scan type: %T", src)
}

// Value returns a sql driver value.
func (uuid UUID) Value() (driver.Value, error) {
	if uuid == nil || len(uuid) == 0 {
		return nil, nil
	}
	return uuid.String(), nil
}
