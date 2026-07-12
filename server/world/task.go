package world

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrWorldClosed means the task's world closed before the task could run.
	ErrWorldClosed = errors.New("world: world closed")
	// ErrEntityClosed means the entity closed before the task could run.
	ErrEntityClosed = errors.New("world: entity closed")
	// ErrEntityNotInWorld means an entity was not in the transaction's world.
	ErrEntityNotInWorld = errors.New("world: entity not in this world")
	// ErrTaskCancelled means the task was cancelled before it started.
	ErrTaskCancelled = errors.New("world: scheduled task cancelled")
	// ErrTaskPanicked means the task's callback panicked; see PanicError.
	ErrTaskPanicked = errors.New("world: scheduled task panicked")
	// ErrEntityType means the entity no longer had the type expected by a
	// typed EntityRef when the task ran.
	ErrEntityType = errors.New("world: unexpected entity type")
)

// PanicError is the Task error for a fire-and-forget callback that panicked. It
// matches errors.Is(err, ErrTaskPanicked) and keeps the original panic value
// and stack. Synchronous Call functions re-panic with Value automatically.
type PanicError struct {
	// Value is the recovered panic value.
	Value any
	// Stack is the stack of the panicking goroutine.
	Stack []byte
}

// Error implements the error interface.
func (e *PanicError) Error() string {
	return fmt.Sprintf("world: scheduled task panicked: %v", e.Value)
}

// Unwrap returns ErrTaskPanicked so errors.Is works.
func (e *PanicError) Unwrap() error { return ErrTaskPanicked }

// RethrowPanic re-panics with the original panic value if an error obtained
// from a Task wraps a *PanicError. The original goroutine's stack remains in
// PanicError.Stack and is logged through the World's Logger when recovered.
// RethrowPanic does nothing for any other error, including nil. Call functions
// invoke it automatically.
func RethrowPanic(err error) {
	if pe, ok := errors.AsType[*PanicError](err); ok {
		panic(pe.Value)
	}
}

// callContext normalises a possibly-nil caller context and reports whether it
// was cancelled before any work was scheduled.
func callContext(ctx context.Context) (context.Context, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx, ctx.Err()
}

// executeWithRecovery runs f, recovering and logging any panic through w.
func executeWithRecovery(w *World, f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			panicErr := &PanicError{Value: r, Stack: debug.Stack()}
			w.conf.Log.Error("scheduled task panicked", "panic", panicErr.Value, "stack", string(panicErr.Stack))
			err = panicErr
		}
	}()
	return f()
}

// awaitTask waits for a task to complete or for cancellation to stop a pending
// task, returning the result stored by the callback through the result pointer.
// Once a callback starts, awaitTask waits for it to finish so synchronous calls
// preserve their result and panic semantics.
func awaitTask[T any](ctx context.Context, task *Task, result *T) (T, error) {
	var zero T
	completed := func() (T, error) {
		if err := task.Err(); err != nil {
			RethrowPanic(err)
			return zero, err
		}
		return *result, nil
	}
	select {
	case <-task.Done():
		return completed()
	case <-ctx.Done():
		if task.Cancel() {
			return zero, ctx.Err()
		}
		<-task.Done()
		return completed()
	}
}

const (
	taskPending int32 = iota
	taskRunning
	taskDone
	taskCancelled
)

// Task tracks work scheduled onto a world or entity owner. Tasks are usually
// fire-and-forget: Done, Err and Wait are for code running off the owner,
// such as tests and shutdown paths. A zero-value Task behaves like a cancelled
// task.
type Task struct {
	done  chan struct{}
	state atomic.Int32

	errMu sync.Mutex
	err   error

	cancelMu sync.Mutex
	onCancel func()
}

// newTask returns a pending Task with an open done channel.
func newTask() *Task {
	return &Task{done: make(chan struct{})}
}

// NewFinishedTask returns a Task that already completed with err.
func NewFinishedTask(err error) *Task {
	t := newTask()
	t.failIfPending(err)
	return t
}

// closedDone is the Done channel returned for nil tasks.
var closedDone = func() <-chan struct{} {
	c := make(chan struct{})
	close(c)
	return c
}()

// Done returns a channel that closes once the task has run, failed or been
// cancelled.
func (t *Task) Done() <-chan struct{} {
	if t == nil || t.done == nil {
		return closedDone
	}
	return t.done
}

// Err returns the task's error, or nil while the task is still pending or
// after it succeeded.
func (t *Task) Err() error {
	if t == nil || t.done == nil {
		return ErrTaskCancelled
	}
	select {
	case <-t.done:
		t.errMu.Lock()
		defer t.errMu.Unlock()
		return t.err
	default:
		return nil
	}
}

// Wait blocks until the task finishes or ctx is cancelled. Never call it from
// a callback running on the same owner: that blocks the owner on itself, the
// deadlock Do exists to avoid.
func (t *Task) Wait(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if t == nil || t.done == nil {
		return ErrTaskCancelled
	}
	select {
	case <-t.done:
		return t.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

// OnDone calls f with the task's error on a fresh goroutine once the task
// completes. Nil and zero-value tasks call f with ErrTaskCancelled. The
// callback always runs on a fresh goroutine, including for those tasks.
func (t *Task) OnDone(f func(err error)) {
	if t == nil || t.done == nil {
		go f(ErrTaskCancelled)
		return
	}
	go func() {
		<-t.done
		f(t.Err())
	}()
}

// Cancel stops a task that has not started yet, reporting whether it did:
// true means the task will never run.
func (t *Task) Cancel() bool {
	if t == nil || t.done == nil || !t.state.CompareAndSwap(taskPending, taskCancelled) {
		return false
	}
	t.setErr(ErrTaskCancelled)
	close(t.done)
	t.runCancel()
	return true
}

// begin moves the task from pending to running, reporting whether it did.
func (t *Task) begin() bool {
	return t != nil && t.state.CompareAndSwap(taskPending, taskRunning)
}

// failIfPending completes a still-pending task with err, reporting whether it
// did.
func (t *Task) failIfPending(err error) bool {
	if t == nil || !t.state.CompareAndSwap(taskPending, taskRunning) {
		return false
	}
	t.finish(err)
	return true
}

// finish completes the task, storing err and closing the done channel.
func (t *Task) finish(err error) {
	t.setErr(err)
	t.state.Store(taskDone)
	close(t.done)
}

func (t *Task) setErr(err error) {
	t.errMu.Lock()
	t.err = err
	t.errMu.Unlock()
}

func (t *Task) pending() bool {
	return t != nil && t.state.Load() == taskPending
}

// setCancel registers a function to run if the task is cancelled. If the
// task is already cancelled when setCancel is called, f runs immediately.
func (t *Task) setCancel(f func()) {
	if t == nil || f == nil {
		return
	}
	t.cancelMu.Lock()
	cancelled := t.state.Load() == taskCancelled
	if !cancelled {
		t.onCancel = f
	}
	t.cancelMu.Unlock()
	if cancelled {
		f()
	}
}

// runCancel invokes the registered cancel function, if any.
func (t *Task) runCancel() {
	t.cancelMu.Lock()
	f := t.onCancel
	t.cancelMu.Unlock()
	if f != nil {
		f()
	}
}

// Do schedules f to run on the world owner and returns immediately; it is
// safe to call from anywhere, including owner callbacks. Tasks usually run in
// submission order, but ordering is not guaranteed between tasks scheduled
// while the queue is saturated. Use one task or Tx.Defer for strictly ordered
// work. On a synchronous World, f runs before Do returns.
func (w *World) Do(f func(tx *Tx)) *Task {
	return w.scheduleTask(newTask(), func(tx *Tx) error {
		f(tx)
		return nil
	})
}

// DoAfter schedules f to run on the world owner after delay. Cancelling the
// task before delay elapses stops f from being queued at all.
func (w *World) DoAfter(delay time.Duration, f func(tx *Tx)) *Task {
	t := newTask()
	run := func(tx *Tx) error {
		f(tx)
		return nil
	}
	if delay <= 0 {
		return w.scheduleTask(t, run)
	}
	if w == nil || w.queue == nil || w.closed.Load() {
		t.failIfPending(ErrWorldClosed)
		return t
	}
	go func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-timer.C:
			w.scheduleTask(t, run)
		case <-t.Done():
		case <-w.closeStarted:
			t.failIfPending(ErrWorldClosed)
		case <-w.closing:
			t.failIfPending(ErrWorldClosed)
		case <-w.queueClosing:
			t.failIfPending(ErrWorldClosed)
		}
	}()
	return t
}

// Call runs f on w's owner and waits for its typed result. It is for
// off-owner code such as tests, startup and background goroutines; if you
// already have a *world.Tx, just use it directly. Calling it from the
// owner itself (any scheduled callback or Handler event) deadlocks. If f
// panics, Call re-panics with the original value on the waiting goroutine after
// logging the original stack through the World's Logger. Context cancellation
// stops pending work, but Call waits for a callback that has already started.
func Call[T any](ctx context.Context, w *World, f func(tx *Tx) (T, error)) (T, error) {
	var zero T
	ctx, err := callContext(ctx)
	if err != nil {
		return zero, err
	}
	var result T
	task := w.scheduleTask(newTask(), func(tx *Tx) error {
		var err error
		result, err = f(tx)
		return err
	})
	return awaitTask(ctx, task, &result)
}

// CallEntity runs f with the EntityHandle's entity on its current world owner
// and waits for the typed result. Off-owner code only, like Call. If f panics,
// CallEntity re-panics with the original value on the waiting goroutine.
func CallEntity[T any](ctx context.Context, h *EntityHandle, f func(tx *Tx, e Entity) (T, error)) (T, error) {
	return CallRef(ctx, NewEntityRef[Entity](h), f)
}

// scheduleTask enqueues a scheduledTransaction on the world's owner queue,
// handing a full queue off to a helper goroutine rather than blocking.
func (w *World) scheduleTask(task *Task, f func(tx *Tx) error) *Task {
	if task == nil {
		task = newTask()
	}
	if w == nil || w.queue == nil || w.closed.Load() {
		task.failIfPending(ErrWorldClosed)
		return task
	}
	if !task.pending() {
		return task
	}
	st := scheduledTransaction{task: task, f: f}
	w.scheduleMu.Lock()
	if w.closed.Load() {
		w.scheduleMu.Unlock()
		task.failIfPending(ErrWorldClosed)
		return task
	}
	if w.conf.Synchronous {
		w.scheduleMu.Unlock()
		st.Run(w)
		return task
	}
	select {
	case <-w.closing:
		task.failIfPending(ErrWorldClosed)
	case <-w.queueClosing:
		task.failIfPending(ErrWorldClosed)
	case w.queue <- st:
	default:
		w.scheduling.Add(1)
		go w.queueScheduled(st)
	}
	w.scheduleMu.Unlock()
	return task
}

// queueScheduled retries enqueuing st once the queue, full at schedule time,
// has room, failing the task if the world closes first.
func (w *World) queueScheduled(st scheduledTransaction) {
	defer w.scheduling.Done()
	if w.closed.Load() {
		st.task.failIfPending(ErrWorldClosed)
		return
	}
	select {
	case <-w.closing:
		st.task.failIfPending(ErrWorldClosed)
	case <-w.queueClosing:
		st.task.failIfPending(ErrWorldClosed)
	case <-st.task.Done():
	case w.queue <- st:
	}
}

// scheduledTransaction is a queued task from Do, DoAfter or Tx.Defer: it
// runs the callback with panic recovery, drains deferred work and finishes the
// task.
type scheduledTransaction struct {
	task *Task
	f    func(tx *Tx) error
}

// Run executes the scheduled callback on the world goroutine.
func (st scheduledTransaction) Run(w *World) {
	if !st.task.begin() {
		return
	}
	tx := newTx(w)
	err := executeWithRecovery(w, func() error { return st.f(tx) })
	tx.close()
	tx.runDeferred()
	st.task.finish(err)
}
