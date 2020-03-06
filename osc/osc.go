package osc

// this package implements a generic online schema change process

// Schema is the subject for osc
// each method should block until the state has been synced, or aborted, or error
type Schema interface {
	EnterAbsent() (bool, error)
	EnterDeleteOnly(add bool) (bool, error)
	EnterWriteOnly(add bool) (bool, error)
	EnterReorgAfterWriteOnly() (bool, error)
	EnterReorgAfterDeleteOnly() (bool, error)
	EnterPublic() (bool, error)
}

// Start osc process
// quit when abort or error, !aborted && err == nil when success
func Start(s Schema, add bool) (aborted bool, err error) {
	switch add {
	case true:
		// add schema

		aborted, err = s.EnterDeleteOnly(add)
		if aborted || err != nil {
			return
		}
		aborted, err = s.EnterWriteOnly(add)
		if aborted || err != nil {
			return
		}
		aborted, err = s.EnterReorgAfterWriteOnly()
		if aborted || err != nil {
			return
		}
		aborted, err = s.EnterPublic()
		if aborted || err != nil {
			return
		}

	case false:

		// delete schema

		aborted, err = s.EnterWriteOnly(add)
		if aborted || err != nil {
			return
		}
		aborted, err = s.EnterDeleteOnly(add)
		if aborted || err != nil {
			return
		}
		aborted, err = s.EnterReorgAfterDeleteOnly()
		if aborted || err != nil {
			return
		}
		aborted, err = s.EnterAbsent()
		if aborted || err != nil {
			return
		}
	}

	return
}
