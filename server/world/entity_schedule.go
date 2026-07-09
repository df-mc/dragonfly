package world

import "time"

// Do schedules f to run with the entity on its current world owner and
// returns immediately. If the entity is in no world yet, the task waits until
// it enters one or the handle closes. The entity passed to f is only valid
// inside f. On a synchronous World, f runs before Do returns.
func (e *EntityHandle) Do(f func(ctx *Context, e Entity)) *Task {
	return e.schedule(func(ctx *Context, e Entity) error {
		f(ctx, e)
		return nil
	})
}

// DoAfter schedules f to run with the entity after delay, following the
// entity if it changes worlds in the meantime.
func (e *EntityHandle) DoAfter(delay time.Duration, f func(ctx *Context, e Entity)) *Task {
	return e.scheduleAfter(delay, func(ctx *Context, e Entity) error {
		f(ctx, e)
		return nil
	})
}

// schedule runs f on the entity's current world owner from a goroutine that
// waits for the entity to be world-bound. It deliberately never enqueues onto
// the current world directly: that would commit to one world and fail instead
// of following the entity when it migrates.
func (e *EntityHandle) schedule(f func(ctx *Context, e Entity) error) *Task {
	task := newTask()
	if e == nil {
		task.failIfPending(ErrEntityClosed)
		return task
	}
	task.setCancel(e.wakeScheduled)
	w := e.trackCloseSchedule(task)
	if !task.pending() {
		return task
	}
	run := func() {
		if w != nil {
			defer w.scheduling.Done()
		}
		e.runScheduled(task, f, w)
	}
	if e.currentWorldSynchronous() {
		run()
	} else {
		go run()
	}
	return task
}

// scheduleAfter runs its own timer loop instead of reusing World.DoAfter:
// the entity may change worlds during the delay, so the loop re-reads the
// current world's signals every iteration.
func (e *EntityHandle) scheduleAfter(delay time.Duration, f func(ctx *Context, e Entity) error) *Task {
	if delay <= 0 {
		return e.schedule(f)
	}
	task := newTask()
	if e == nil {
		task.failIfPending(ErrEntityClosed)
		return task
	}
	task.setCancel(e.wakeScheduled)
	go func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		for {
			closeStarted, worldChanged := e.currentWorldSignals()
			select {
			case <-timer.C:
				if e.currentWorldClosing() {
					task.failIfPending(ErrWorldClosed)
					return
				}
				e.runScheduled(task, f, nil)
				return
			case <-task.Done():
				return
			case <-closeStarted:
				if cs, _ := e.currentWorldSignals(); cs == closeStarted {
					task.failIfPending(ErrWorldClosed)
					return
				}
			case <-worldChanged:
			case <-e.closed:
				task.failIfPending(ErrEntityClosed)
				return
			}
		}
	}()
	return task
}

// trackCloseSchedule registers entity work created during the world's close
// transaction, so World.close drains it before shutting the queue down.
func (e *EntityHandle) trackCloseSchedule(task *Task) *World {
	e.cond.L.Lock()
	defer e.cond.L.Unlock()
	w := e.w
	if w == nil || w == closeWorld || !w.closed.Load() {
		return nil
	}
	w.scheduleMu.Lock()
	defer w.scheduleMu.Unlock()
	if !w.closeAcceptingEntityTasks.Load() {
		task.failIfPending(ErrWorldClosed)
		return nil
	}
	select {
	case <-w.queueClosing:
		task.failIfPending(ErrWorldClosed)
		return nil
	default:
		w.scheduling.Add(1)
		return w
	}
}

// wakeScheduled wakes goroutines waiting on the handle's cond, so a scheduler
// blocked in execWorld re-checks its cancel signal.
func (e *EntityHandle) wakeScheduled() {
	e.cond.L.Lock()
	e.cond.Broadcast()
	e.cond.L.Unlock()
}

// currentWorldSignals returns the current world's close channel and the
// handle's world-change channel, creating the latter for this waiter if
// needed.
func (e *EntityHandle) currentWorldSignals() (<-chan struct{}, <-chan struct{}) {
	e.cond.L.Lock()
	defer e.cond.L.Unlock()
	if e.worldChanged == nil {
		e.worldChanged = make(chan struct{})
	}
	if e.w == nil || e.w == closeWorld {
		return nil, e.worldChanged
	}
	return e.w.closeStarted, e.worldChanged
}

// currentWorldSynchronous reports whether the entity is bound to a
// synchronous World.
func (e *EntityHandle) currentWorldSynchronous() bool {
	e.cond.L.Lock()
	defer e.cond.L.Unlock()
	return e.w != nil && e.w != closeWorld && e.worldReady && e.w.conf.Synchronous
}

// currentWorldClosing reports whether the entity's current world has started
// closing.
func (e *EntityHandle) currentWorldClosing() bool {
	closeStarted, _ := e.currentWorldSignals()
	return cancelled(closeStarted)
}

// runScheduled executes the scheduled entity callback via execWorld with the
// same completion model as scheduledTransaction: run, drain deferred work,
// finish the task.
func (e *EntityHandle) runScheduled(task *Task, f func(ctx *Context, e Entity) error, allowedCloseWorld *World) {
	run := e.execWorld(func(ctx *Context, ent Entity) {
		if !task.begin() {
			return
		}
		err := executeWithRecovery(ctx.w, func() error { return f(ctx, ent) })
		ctx.runDeferred()
		task.finish(err)
	}, false, task.Done(), allowedCloseWorld)
	if !run || task.pending() {
		err := ErrEntityClosed
		if e.currentWorldClosing() {
			err = ErrWorldClosed
		}
		task.failIfPending(err)
	}
}
