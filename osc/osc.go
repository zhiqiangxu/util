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

// SchemaChange is the subject for osc
// each method should block until the state has been synced, or error
type SchemaChange interface {
	GetState() SchemaState
	EnterAbsent() error
	EnterDeleteOnly(add bool) error
	EnterWriteOnly(add bool) error
	EnterReorgAfterWriteOnly() error
	EnterReorgAfterDeleteOnly() error
	EnterPublic() error
}

const (
	errFormat = "state not expected, got %d should %d"
)

// Start osc process
func Start(s SchemaChange, add bool) (err error) {
	switch add {
	case true:
		// add schema

		state := s.GetState()
		if state != StateAbsent {
			err = fmt.Errorf(errFormat, state, StateAbsent)
			return
		}
		err = s.EnterDeleteOnly(add)
		if err != nil {
			return
		}
		state = s.GetState()
		if state != StateDeleteOnly {
			err = fmt.Errorf(errFormat, state, StateDeleteOnly)
			return
		}
		err = s.EnterWriteOnly(add)
		if err != nil {
			return
		}
		state = s.GetState()
		if state != StateWriteOnly {
			err = fmt.Errorf(errFormat, state, StateWriteOnly)
			return
		}
		err = s.EnterReorgAfterWriteOnly()
		if err != nil {
			return
		}
		state = s.GetState()
		if state != StateWriteReorganization {
			err = fmt.Errorf(errFormat, state, StateWriteReorganization)
			return
		}
		err = s.EnterPublic()
		if err != nil {
			return
		}
		state = s.GetState()
		if state != StatePublic {
			err = fmt.Errorf(errFormat, state, StatePublic)
			return
		}

	case false:

		// delete schema

		state := s.GetState()
		if state != StatePublic {
			err = fmt.Errorf(errFormat, state, StatePublic)
			return
		}

		err = s.EnterWriteOnly(add)
		if err != nil {
			return
		}
		if state != StateWriteOnly {
			err = fmt.Errorf(errFormat, state, StateWriteOnly)
			return
		}
		err = s.EnterDeleteOnly(add)
		if err != nil {
			return
		}
		if state != StateDeleteOnly {
			err = fmt.Errorf(errFormat, state, StateDeleteOnly)
			return
		}
		err = s.EnterReorgAfterDeleteOnly()
		if err != nil {
			return
		}
		if state != StateDeleteReorganization {
			err = fmt.Errorf(errFormat, state, StateDeleteReorganization)
			return
		}
		err = s.EnterAbsent()
		if err != nil {
			return
		}
		if state != StateAbsent {
			err = fmt.Errorf(errFormat, state, StateAbsent)
			return
		}
	}

	return
}
