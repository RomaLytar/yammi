package usecase

import "time"

func getCurrentTime() time.Time {
	return time.Now().UTC()
}
