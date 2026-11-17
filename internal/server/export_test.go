package server

import "time"

// SetStopTime sets the longest time of the stop of s.
func SetStopTime(s *Server, d time.Duration) {
	s.stopTime = d
}
