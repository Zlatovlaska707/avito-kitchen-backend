package nower

import "time"

type Real struct{}

func NewReal() Real {
	return Real{}
}

func (Real) Now() time.Time {
	return time.Now().UTC()
}

type Fixed struct {
	t time.Time
}

func NewFixed(t time.Time) Fixed {
	return Fixed{t: t.UTC()}
}

func (f Fixed) Now() time.Time {
	return f.t
}
