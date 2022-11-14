package reconciler

type state int

const (
	updated state = iota
	skipped
	failed
)

type Result struct {
	s state
	e error
}

func newUpdatedResult() Result {
	return Result{s: updated, e: nil}
}

func newSkippedResult() Result {
	return Result{s: skipped, e: nil}
}

func newFailedResult(e error) Result {
	return Result{s: failed, e: e}
}

func (r Result) Updated() bool {
	return r.s == updated
}

func (r Result) Skipped() bool {
	return r.s == skipped
}

func (r Result) Failed() bool {
	return r.s == failed
}

func (r Result) GetError() error {
	return r.e
}
