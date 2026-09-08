package nower

import "time"

type Nower interface {
	Now() time.Time
}
