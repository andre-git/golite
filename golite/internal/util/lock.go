package util

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrBusy = errors.New("database is busy")
)

type LockLevel int

const (
	None      LockLevel = 0
	Shared    LockLevel = 1
	Reserved  LockLevel = 2
	Pending   LockLevel = 3
	Exclusive LockLevel = 4
)

func (l LockLevel) String() string {
	switch l {
	case None: return "NONE"
	case Shared: return "SHARED"
	case Reserved: return "RESERVED"
	case Pending: return "PENDING"
	case Exclusive: return "EXCLUSIVE"
	default: return fmt.Sprintf("UNKNOWN(%d)", l)
	}
}

type LockManager struct {
	rw    sync.RWMutex 
	resMu sync.Mutex   
	penMu sync.Mutex   

	stateMu sync.Mutex 
	cond    *sync.Cond
}

func NewLockManager() *LockManager {
	lm := &LockManager{}
	lm.cond = sync.NewCond(&lm.stateMu)
	return lm
}

func (lm *LockManager) TryLock(current *LockLevel, target LockLevel) error {
	if target <= *current {
		return nil
	}

	switch target {
	case Shared:
		if !lm.penMu.TryLock() {
			return ErrBusy
		}
		if !lm.rw.TryRLock() {
			lm.penMu.Unlock()
			return ErrBusy
		}
		lm.penMu.Unlock()
		*current = Shared
		return nil

	case Reserved:
		if *current < Shared {
			return fmt.Errorf("must hold SHARED lock to acquire RESERVED")
		}
		if !lm.resMu.TryLock() {
			return ErrBusy
		}
		*current = Reserved
		return nil

	case Exclusive:
		if *current < Shared {
			return fmt.Errorf("must hold SHARED lock to acquire EXCLUSIVE")
		}

		if *current < Pending {
			if !lm.penMu.TryLock() {
				return ErrBusy
			}
			*current = Pending
		}

		lm.rw.RUnlock()
		if !lm.rw.TryLock() {
			lm.rw.RLock()
			return ErrBusy
		}
		*current = Exclusive
		return nil
	}

	return fmt.Errorf("invalid lock transition to %s", target)
}

func (lm *LockManager) Unlock(current *LockLevel, target LockLevel) error {
	if target >= *current {
		return nil
	}

	if *current == Exclusive {
		lm.rw.Unlock()
	} else if *current >= Shared {
		lm.rw.RUnlock()
	}

	if *current >= Pending {
		lm.penMu.Unlock()
	}

	if *current >= Reserved {
		lm.resMu.Unlock()
	}

	if target == Shared {
		lm.rw.RLock()
	}

	*current = target
	lm.cond.Broadcast() 
	return nil
}

func (lm *LockManager) CheckReserved() bool {
	if lm.resMu.TryLock() {
		lm.resMu.Unlock()
		return false
	}
	return true
}
