package monitor

import "testing"

func TestService_Info(t *testing.T) {
	s := &Service{}
	s.Info()
}

func TestService_Stats(t *testing.T) {
	s := &Service{}
	s.Stats()
}
