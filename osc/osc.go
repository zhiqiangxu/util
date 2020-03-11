package osc

import "fmt"

// this package implements a generic online schema change process

// SchemaState is the state for schema elements.
type SchemaState byte

const (
	// StateAbsent means this schema element is absent and can't be used.
	StateAbsent SchemaState = iota
	// StateDeleteOnly means we can only delete items for this schema element.
	StateDeleteOnly
	// StateWriteOnly means we can use any write operation on this schema element,
	// but outer can't read the changed data.
	StateWriteOnly
	// StateWriteReorganization means we are re-organizing whole data after write only state.
	StateWriteReorganization
	// StateDeleteReorganization means we are re-organizing whole data after delete only state.
	StateDeleteReorganization
	// StatePublic means this schema element is ok for all write and read operations.
	StatePublic
)

// AddSchemaChange for add schema change
// each method should block until the state has been synced, or error
type AddSchemaChange interface {
	GetState() SchemaState
	EnterDeleteOnly() error
	EnterWriteOnly() error
	EnterReorgAfterWriteOnly() error
	EnterPublic() error
}

// DeleteSchemaChange for delete schema change
// each method should block until the state has been synced, or error
type DeleteSchemaChange interface {
	GetState() SchemaState
	EnterWriteOnly() error
	EnterDeleteOnly() error
	EnterReorgAfterDeleteOnly() error
	EnterAbsent() error
}

const (
	errInvalidState = "invalid state %d"
)

// StepAdd for a single add schema state transition
func StepAdd(asc AddSchemaChange) (err error) {
	state := asc.GetState()
	switch state {
	case StateAbsent:
		err = asc.EnterDeleteOnly()
		return
	case StateDeleteOnly:
		err = asc.EnterWriteOnly()
		return
	case StateWriteOnly:
		err = asc.EnterReorgAfterWriteOnly()
		return
	case StateWriteReorganization:
		err = asc.EnterPublic()
		return
	case StatePublic:
		return
	default:
		err = fmt.Errorf(errInvalidState, state)
		return
	}
}

// StepDelete for a single delete schema state transition
func StepDelete(dsc DeleteSchemaChange) (err error) {
	state := dsc.GetState()
	switch state {
	case StatePublic:
		err = dsc.EnterWriteOnly()
		return
	case StateWriteOnly:
		err = dsc.EnterDeleteOnly()
		return
	case StateDeleteOnly:
		err = dsc.EnterReorgAfterDeleteOnly()
		return
	case StateDeleteReorganization:
		err = dsc.EnterAbsent()
		return
	case StateAbsent:
		return
	default:
		err = fmt.Errorf(errInvalidState, state)
		return
	}
}

// StartAdd for start osc add process
func StartAdd(asc AddSchemaChange) (err error) {
	for asc.GetState() != StatePublic {
		err = StepAdd(asc)
		if err != nil {
			return
		}
	}
	return
}

// StartDelete for start osc delete process
func StartDelete(dsc DeleteSchemaChange) (err error) {
	for dsc.GetState() != StateAbsent {
		err = StepDelete(dsc)
		if err != nil {
			return
		}
	}
	return
}
