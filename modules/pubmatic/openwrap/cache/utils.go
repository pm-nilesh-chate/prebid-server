package cache

import "time"

func getSeconds(duration int) time.Duration {
	return time.Duration(duration) * time.Second
}
