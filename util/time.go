package util

import "time"

func TimeInMilliseconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
