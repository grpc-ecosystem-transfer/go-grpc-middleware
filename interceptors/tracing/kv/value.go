package kv

import (
	"math"
)

// KeyValue holds a key and value pair.
type KeyValue struct {
	Key   Key
	Value Value
}

type Key string

// Bool creates a KeyValue instance with a BOOL Value.
//
// If creating both key and a bool value at the same time, then
// instead of calling kv.Key(name).Bool(value) consider using a
// convenience function provided by the api/key package -
// key.Bool(name, value).
func (k Key) Bool(v bool) KeyValue {
	return KeyValue{
		Key:   k,
		Value: boolValue(v),
	}
}

// Int64 creates a KeyValue instance with an INT64 Value.
//
// If creating both key and an int64 value at the same time, then
// instead of calling kv.Key(name).Int64(value) consider using a
// convenience function provided by the api/key package -
// key.Int64(name, value).
func (k Key) Int64(v int64) KeyValue {
	return KeyValue{
		Key:   k,
		Value: int64Value(v),
	}
}

// Float64 creates a KeyValue instance with a FLOAT64 Value.
//
// If creating both key and a float64 value at the same time, then
// instead of calling kv.Key(name).Float64(value) consider using a
// convenience function provided by the api/key package -
// key.Float64(name, value).
func (k Key) Float64(v float64) KeyValue {
	return KeyValue{
		Key:   k,
		Value: float64Value(v),
	}
}

// String creates a KeyValue instance with a STRING Value.
//
// If creating both key and a string value at the same time, then
// instead of calling kv.Key(name).String(value) consider using a
// convenience function provided by the api/key package -
// key.String(name, value).
func (k Key) String(v string) KeyValue {
	return KeyValue{
		Key:   k,
		Value: stringValue(v),
	}
}

// Int creates a KeyValue instance with either an INT32 or an INT64
// Value, depending on whether the int type is 32 or 64 bits wide.
//
// If creating both key and an int value at the same time, then
// instead of calling kv.Key(name).Int(value) consider using a
// convenience function provided by the api/key package -
// key.Int(name, value).
func (k Key) Int(v int) KeyValue {
	return KeyValue{
		Key:   k,
		Value: intValue(v),
	}
}

// ValueType describes the type of the data Value holds.
type ValueType int

const (
	INVALID ValueType = iota // No value.
	// BOOL is a boolean Type Value.
	BOOL
	// INT64 is a 64-bit signed integral Type Value.
	INT64
	// FLOAT64 is a 64-bit floating point Type Value.
	FLOAT64
	// STRING is a string Type Value.
	STRING
)

type Value struct {
	vtype    ValueType
	numeric  uint64
	stringly string
}

func boolTowRaw(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}

func rawToBool(r uint64) bool {
	return r != 0
}

func boolValue(v bool) Value {
	return Value{
		vtype:   BOOL,
		numeric: boolTowRaw(v),
	}
}

func int64Value(v int64) Value {
	return Value{
		vtype:   INT64,
		numeric: uint64(v),
	}
}

func float64Value(v float64) Value {
	return Value{
		vtype:   FLOAT64,
		numeric: math.Float64bits(v),
	}
}

func stringValue(v string) Value {
	return Value{
		vtype:    STRING,
		stringly: v,
	}
}

// intValue creates an INT64 Value.
func intValue(v int) Value {
	return int64Value(int64(v))
}

func (v Value) AsBool() bool {
	return rawToBool(v.numeric)
}

// AsInt64 returns the int64 value. Make sure that the Value's type is
// INT64.
func (v Value) AsInt64() int64 {
	return int64(v.numeric)
}

// AsFloat64 returns the float64 value. Make sure that the Value's
// type is FLOAT64.
func (v Value) AsFloat64() float64 {
	return math.Float64frombits(v.numeric)
}

// AsString returns the string value. Make sure that the Value's type
// is STRING.
func (v Value) AsString() string {
	return v.stringly
}

// Type returns a type of the Value.
func (v Value) Type() ValueType {
	return v.vtype
}
