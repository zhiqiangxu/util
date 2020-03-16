package deadlock

import (
	"runtime"

	"github.com/petermattis/goid"
	"github.com/zhiqiangxu/util"
)

type detector struct {
	ownerResouces map[int64]map[uint64]bool
	resouceOwners map[uint64]*resourceOwner
	waitForMap    map[int64]*waitForResource
}

type resourceOwner struct {
	wgid  int64
	rgids map[int64]int
}

type waitForResource struct {
	resourceID uint64
	w          bool
}

func newDetector() *detector {
	return &detector{
		ownerResouces: make(map[int64]map[uint64]bool),
		resouceOwners: make(map[uint64]*resourceOwner),
		waitForMap:    make(map[int64]*waitForResource),
	}
}

func (d *detector) onAcquiredLocked(resourceID uint64, w bool) {
	gid := goid.Get()

	// update ownerResouces
	ownedResources := d.ownerResouces[gid]
	if ownedResources == nil {
		ownedResources = make(map[uint64]bool)
		d.ownerResouces[gid] = ownedResources
	}
	ownedResources[resourceID] = w

	// update resouceOwners
	resourceOwners := d.resouceOwners[resourceID]
	if resourceOwners == nil {
		resourceOwners = &resourceOwner{}
		d.resouceOwners[resourceID] = resourceOwners
	}
	if w {
		if resourceOwners.wgid != 0 {
			panic("write lock holding by more than one owners")
		}
		resourceOwners.wgid = gid
	} else {
		rgids := resourceOwners.rgids
		if rgids == nil {
			rgids = make(map[int64]int)
			resourceOwners.rgids = rgids
		}
		rgids[gid]++
	}

	// update waitForMap
	delete(d.waitForMap, gid)
}

// ErrorDeadlock contains deadlock info
type ErrorDeadlock struct {
	SourceParty Party
	OwnerParty  Party
	Stack       string
}

// ErrorUsage for incorrect lock usage
type ErrorUsage struct {
	Msg   string
	Stack string
}

// Party for one side of deadlock
type Party struct {
	GID        int64
	ResourceID uint64
	W          bool
}

func getCallStack() string {
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:runtime.Stack(buf, false)]
	return util.String(buf)
}

// ParsePanicError returns non nil ErrorDeadlock if deadlock happend
// the ErrorUsage is non nil for lock usage problems
func ParsePanicError(panicErr interface{}) (edl *ErrorDeadlock, errUsage *ErrorUsage) {
	if panicErr == nil {
		return
	}

	if panicErrStr, ok := panicErr.(string); ok {
		errUsage = &ErrorUsage{
			Msg:   panicErrStr,
			Stack: getCallStack(),
		}
		return
	}

	if panicErrDL, ok := panicErr.(*ErrorDeadlock); ok {
		panicErrDL.Stack = getCallStack()
		edl = panicErrDL
		return
	}

	panic("bug happened")
}

func (d *detector) onWaitLocked(resourceID uint64, w bool) {
	gid := goid.Get()

	if d.waitForMap[gid] != nil {
		panic("waiting for multiple resources")
	}

	resourceOwners := d.resouceOwners[resourceID]
	if resourceOwners == nil {
		panic("waiting for a resource with no owner")
	}
	if resourceOwners.wgid == 0 && len(resourceOwners.rgids) == 0 {
		panic("waiting for a resource with no owner")
	}

	// detect deadlock
	var err *ErrorDeadlock
	// check deadlock with write lock owner
	if resourceOwners.wgid != 0 {
		err = d.doDetect(gid, resourceOwners.wgid)
		if err != nil {
			err.OwnerParty = Party{GID: resourceOwners.wgid, ResourceID: resourceID, W: true}
			panic(err)
		}
	}
	// check deadlock with read lock owner
	for rgid := range resourceOwners.rgids {
		err = d.doDetect(gid, rgid)
		if err != nil {
			err.OwnerParty = Party{GID: rgid, ResourceID: resourceID, W: false}
			panic(err)
		}
	}

	d.waitForMap[gid] = &waitForResource{resourceID: resourceID, w: w}
}

func (d *detector) doDetect(sourceGID, ownerGID int64) (err *ErrorDeadlock) {
	waitingForResource := d.waitForMap[ownerGID]
	if waitingForResource == nil {
		return
	}

	resourceOwners := d.resouceOwners[waitingForResource.resourceID]
	if resourceOwners == nil || (resourceOwners.wgid == 0 && len(resourceOwners.rgids) == 0) {
		panic("waiting for a resource with no owner")
	}

	if resourceOwners.wgid != 0 {
		if resourceOwners.wgid == sourceGID {
			err = &ErrorDeadlock{SourceParty: Party{
				GID:        sourceGID,
				ResourceID: waitingForResource.resourceID,
				W:          true,
			}}
			return
		}
		err = d.doDetect(sourceGID, resourceOwners.wgid)
		if err != nil {
			return
		}
	}

	for rgid := range resourceOwners.rgids {
		if rgid == sourceGID {
			err = &ErrorDeadlock{
				SourceParty: Party{
					GID:        sourceGID,
					ResourceID: waitingForResource.resourceID,
					W:          false,
				}}
			return
		}
		err = d.doDetect(sourceGID, rgid)
		if err != nil {
			return
		}
	}
	return
}

func (d *detector) onReleaseLocked(resourceID uint64, w bool) {
	gid := goid.Get()

	// update ownerResouces
	ownedResources := d.ownerResouces[gid]
	if ownedResources == nil {
		panic("releasing a lock not owned")
	}
	if _, exists := ownedResources[resourceID]; !exists {
		panic("releasing a lock not owned")
	}
	delete(ownedResources, resourceID)
	if len(ownedResources) == 0 {
		delete(d.ownerResouces, gid)
	}

	// update resouceOwners
	resourceOwners := d.resouceOwners[resourceID]
	if resourceOwners == nil {
		panic("releasing a lock not owned")
	}
	if w {
		if resourceOwners.wgid != gid {
			panic("releasing a lock not owned")
		}
		resourceOwners.wgid = 0
		if len(resourceOwners.rgids) == 0 {
			delete(d.resouceOwners, resourceID)
		}
	} else {
		if _, exists := resourceOwners.rgids[gid]; !exists {
			panic("releasing a lock not owned")
		}

		resourceOwners.rgids[gid]--
		if resourceOwners.rgids[gid] == 0 {
			delete(resourceOwners.rgids, gid)
			if len(resourceOwners.rgids) == 0 && resourceOwners.wgid == 0 {
				delete(d.resouceOwners, resourceID)
			}
		} else if resourceOwners.rgids[gid] < 0 {
			panic("releasing a read lock too many times")
		}
	}
}

var d *detector

func init() {
	d = newDetector()
}
