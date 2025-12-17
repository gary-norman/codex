package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type AnyID interface {
	UUIDField | int64
}

type UUIDField struct {
	UUID uuid.UUID
}

// NullableUUIDField is a nullable UUID type
type NullableUUIDField struct {
	UUID  UUIDField
	Valid bool // true if UUID is not NULL
}

// NewUUIDField automatically generates a new UUID if it's not already set
func NewUUIDField() UUIDField {
	return UUIDField{UUID: uuid.New()}
}

// ZeroUUIDField returns a UUIDField with a nil UUID
func ZeroUUIDField() UUIDField {
	return UUIDField{UUID: uuid.Nil}
}

// -------------------------------------------------------
// Methods for JSON marshaling/unmarshaling
// -------------------------------------------------------

// String implements fmt.Stringer
func (u UUIDField) String() string {
	return u.UUID.String()
}

// MarshalJSON marshals the UUID string as JSON
func (u UUIDField) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.UUID.String())
}

// UnmarshalJSON unmarshals the UUID from JSON
func (u *UUIDField) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	parsed, err := uuid.Parse(str)
	if err != nil {
		return err
	}
	u.UUID = parsed
	return nil
}

// -------------------------------------------------------
// SQL driver interfaces
// -------------------------------------------------------

// copyFromBytes copies the UUID from a byte slice.
func (u *UUIDField) copyFromBytes(src any) error {
	switch v := src.(type) {
	case []byte:
		copy(u.UUID[:], v)
		return nil
	default:
		return fmt.Errorf("UUIDField: cannot scan type %T", v)
	}
}

// Exec implements the sql.Execer interface.
func (u *UUIDField) Exec(value any) error {
	return u.copyFromBytes(value)
}

// Scan implements the sql.Scanner interface.
func (u *UUIDField) Scan(value any) error {
	return u.copyFromBytes(value)
}

// Begin implements the sql.Tx interface.
func (u *UUIDField) Begin(value any) error {
	return u.copyFromBytes(value)
}

// Commit implements the sql.Tx interface.
func (u *UUIDField) Commit(value any) error {
	return u.copyFromBytes(value)
}

// Value implements the driver.Valuer interface.
func (u UUIDField) Value() (driver.Value, error) {
	return u.UUID[:], nil // store as []byte
}

// UUIDFieldFromString converts a string to a UUIDField
func UUIDFieldFromString(s string) (UUIDField, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return UUIDField{}, err
	}
	return UUIDField{UUID: parsed}, nil
}

// Scan implements the sql.Scanner interface.
func (u *NullableUUIDField) Scan(value any) error {
	if value == nil {
		u.Valid = false
		return nil
	}
	// Use UUIDField's Scan method for non-nil values
	err := u.UUID.Scan(value)
	if err != nil {
		return err
	}
	u.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (u NullableUUIDField) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}
	return u.UUID.Value()
}
