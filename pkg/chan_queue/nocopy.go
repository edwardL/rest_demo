package chan_queue

// noCopy is used to ensure that we don't copy things that shouldn't
// be copied.
//
// See https://golang.org/issues/8005#issuecomment-190753527.
type noCopy struct{}

func (noCopy) Lock() {}

func (noCopy) UnLock() {}
