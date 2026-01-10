package omnillm

import (
	"testing"
	"time"
)

func TestCircuitBreaker_InitialState(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	if cb.State() != CircuitClosed {
		t.Errorf("expected initial state to be closed, got %v", cb.State())
	}

	if !cb.AllowRequest() {
		t.Error("expected AllowRequest to return true in closed state")
	}
}

func TestCircuitBreaker_OpensAfterConsecutiveFailures(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold:     3,
		SuccessThreshold:     2,
		Timeout:              1 * time.Second,
		FailureRateThreshold: 0.5,
		MinimumRequests:      10,
	}
	cb := NewCircuitBreaker(config)

	// Record failures up to threshold
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.State() != CircuitOpen {
		t.Errorf("expected circuit to be open after %d failures, got %v", 3, cb.State())
	}

	if cb.AllowRequest() {
		t.Error("expected AllowRequest to return false in open state")
	}
}

func TestCircuitBreaker_OpensOnFailureRate(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold:     100, // High threshold so we test rate
		SuccessThreshold:     2,
		Timeout:              1 * time.Second,
		FailureRateThreshold: 0.5,
		MinimumRequests:      10,
	}
	cb := NewCircuitBreaker(config)

	// Record 5 successes and 5 failures (50% failure rate)
	for i := 0; i < 5; i++ {
		cb.RecordSuccess()
	}
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	// Should now be open due to 50% failure rate at 10 requests
	if cb.State() != CircuitOpen {
		t.Errorf("expected circuit to be open after 50%% failure rate, got %v", cb.State())
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          50 * time.Millisecond,
		MinimumRequests:  10,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Fatalf("expected circuit to be open, got %v", cb.State())
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Should transition to half-open on next AllowRequest
	if !cb.AllowRequest() {
		t.Error("expected AllowRequest to return true after timeout")
	}

	if cb.State() != CircuitHalfOpen {
		t.Errorf("expected circuit to be half-open, got %v", cb.State())
	}
}

func TestCircuitBreaker_ClosesAfterSuccessesInHalfOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          50 * time.Millisecond,
		MinimumRequests:  10,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Wait for timeout and transition to half-open
	time.Sleep(60 * time.Millisecond)
	cb.AllowRequest()

	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected circuit to be half-open, got %v", cb.State())
	}

	// Record successes to close
	cb.RecordSuccess()
	cb.RecordSuccess()

	if cb.State() != CircuitClosed {
		t.Errorf("expected circuit to be closed after successes, got %v", cb.State())
	}
}

func TestCircuitBreaker_ReopensOnFailureInHalfOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          50 * time.Millisecond,
		MinimumRequests:  10,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Wait for timeout and transition to half-open
	time.Sleep(60 * time.Millisecond)
	cb.AllowRequest()

	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected circuit to be half-open, got %v", cb.State())
	}

	// Record a failure - should reopen
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Errorf("expected circuit to reopen after failure in half-open, got %v", cb.State())
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          1 * time.Second,
		MinimumRequests:  10,
	}
	cb := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Fatalf("expected circuit to be open, got %v", cb.State())
	}

	// Reset
	cb.Reset()

	if cb.State() != CircuitClosed {
		t.Errorf("expected circuit to be closed after reset, got %v", cb.State())
	}

	stats := cb.Stats()
	if stats.TotalRequests != 0 || stats.TotalFailures != 0 {
		t.Errorf("expected counters to be reset, got requests=%d, failures=%d",
			stats.TotalRequests, stats.TotalFailures)
	}
}

func TestCircuitBreaker_Stats(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	// Record some activity
	cb.RecordSuccess()
	cb.RecordSuccess()
	cb.RecordFailure()

	stats := cb.Stats()

	if stats.TotalRequests != 3 {
		t.Errorf("expected 3 total requests, got %d", stats.TotalRequests)
	}

	if stats.TotalFailures != 1 {
		t.Errorf("expected 1 total failure, got %d", stats.TotalFailures)
	}

	expectedRate := 1.0 / 3.0
	if stats.FailureRate < expectedRate-0.01 || stats.FailureRate > expectedRate+0.01 {
		t.Errorf("expected failure rate ~%.2f, got %.2f", expectedRate, stats.FailureRate)
	}

	if stats.ConsecutiveFailures != 1 {
		t.Errorf("expected 1 consecutive failure, got %d", stats.ConsecutiveFailures)
	}
}

func TestCircuitBreaker_DefaultConfig(t *testing.T) {
	config := DefaultCircuitBreakerConfig()

	if config.FailureThreshold != 5 {
		t.Errorf("expected default FailureThreshold=5, got %d", config.FailureThreshold)
	}
	if config.SuccessThreshold != 2 {
		t.Errorf("expected default SuccessThreshold=2, got %d", config.SuccessThreshold)
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("expected default Timeout=30s, got %v", config.Timeout)
	}
	if config.FailureRateThreshold != 0.5 {
		t.Errorf("expected default FailureRateThreshold=0.5, got %f", config.FailureRateThreshold)
	}
	if config.MinimumRequests != 10 {
		t.Errorf("expected default MinimumRequests=10, got %d", config.MinimumRequests)
	}
}

func TestCircuitState_String(t *testing.T) {
	tests := []struct {
		state    CircuitState
		expected string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("CircuitState(%d).String() = %q, want %q", tt.state, got, tt.expected)
		}
	}
}

func TestCircuitOpenError(t *testing.T) {
	err := &CircuitOpenError{
		Provider:   "openai",
		State:      CircuitOpen,
		RetryAfter: 30 * time.Second,
	}

	expected := "circuit breaker is open for provider openai; retry after 30s"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}
