package testutil

import "fmt"

type RequestStep struct {
	Result map[string]interface{}
	Err    error
}

type RetryScript struct {
	steps []RequestStep

	RequestCalls    int
	RefreshCalls    int
	InvalidateCalls int
	ReadDelayCalls  int
	WriteDelayCalls int

	RefreshErr error
}

func NewRetryScript(steps []RequestStep) *RetryScript {
	return &RetryScript{steps: steps}
}

func (s *RetryScript) RequestHook(method, urlStr string, params, jsonData map[string]interface{}, maxRetries int, referer, operation string) (map[string]interface{}, error) {
	s.RequestCalls++
	idx := s.RequestCalls - 1
	if idx < 0 || idx >= len(s.steps) {
		return nil, fmt.Errorf("no scripted request step for call %d", s.RequestCalls)
	}
	return s.steps[idx].Result, s.steps[idx].Err
}

func (s *RetryScript) RefreshHook() error {
	s.RefreshCalls++
	return s.RefreshErr
}

func (s *RetryScript) InvalidateHook() {
	s.InvalidateCalls++
}

func (s *RetryScript) ReadDelayHook() {
	s.ReadDelayCalls++
}

func (s *RetryScript) WriteDelayHook() {
	s.WriteDelayCalls++
}
