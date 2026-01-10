package omnillm

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// CircuitClosed indicates normal operation - requests pass through
	CircuitClosed CircuitState = iota
	// CircuitOpen indicates the circuit is open - requests fail fast
	CircuitOpen
	// CircuitHalfOpen indicates the circuit is testing recovery
	CircuitHalfOpen
)

// String returns the string representation of the circuit state
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig configures circuit breaker behavior
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures before opening the circuit.
	// Default: 5
	FailureThreshold int

	// SuccessThreshold is the number of consecutive successes in half-open state
	// required to close the circuit.
	// Default: 2
	SuccessThreshold int

	// Timeout is how long to wait in open state before transitioning to half-open.
	// Default: 30 seconds
	Timeout time.Duration

	// FailureRateThreshold triggers circuit open when the failure rate exceeds this value (0-1).
	// Only evaluated after MinimumRequests is reached.
	// Default: 0.5 (50%)
	FailureRateThreshold float64

	// MinimumRequests is the minimum number of requests before failure rate is evaluated.
	// Default: 10
	MinimumRequests int
}

// DefaultCircuitBreakerConfig returns a CircuitBreakerConfig with sensible defaults
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold:     5,
		SuccessThreshold:     2,
		Timeout:              30 * time.Second,
		FailureRateThreshold: 0.5,
		MinimumRequests:      10,
	}
}

// CircuitBreaker implements the circuit breaker pattern for provider health tracking
type CircuitBreaker struct {
	mu sync.RWMutex

	config CircuitBreakerConfig
	state  CircuitState

	// Counters for current window
	consecutiveFailures  int
	consecutiveSuccesses int

	// Counters for failure rate calculation
	totalRequests int
	totalFailures int

	// Timing
	lastFailure     time.Time
	lastStateChange time.Time
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration.
// If config has zero values, defaults are used for those fields.
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	// Apply defaults for zero values
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5
	}
	if config.SuccessThreshold == 0 {
		config.SuccessThreshold = 2
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.FailureRateThreshold == 0 {
		config.FailureRateThreshold = 0.5
	}
	if config.MinimumRequests == 0 {
		config.MinimumRequests = 10
	}

	return &CircuitBreaker{
		config:          config,
		state:           CircuitClosed,
		lastStateChange: time.Now(),
	}
}

// AllowRequest returns true if the request should be allowed to proceed.
// In closed state, always allows. In open state, allows only after timeout.
// In half-open state, allows a limited number of test requests.
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true

	case CircuitOpen:
		// Check if timeout has elapsed
		if time.Since(cb.lastFailure) >= cb.config.Timeout {
			cb.transitionTo(CircuitHalfOpen)
			return true
		}
		return false

	case CircuitHalfOpen:
		// Allow test requests in half-open state
		return true

	default:
		return true
	}
}

// RecordSuccess records a successful request.
// In half-open state, may close the circuit if enough successes.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalRequests++
	cb.consecutiveFailures = 0
	cb.consecutiveSuccesses++

	switch cb.state {
	case CircuitHalfOpen:
		// Check if we have enough consecutive successes to close
		if cb.consecutiveSuccesses >= cb.config.SuccessThreshold {
			cb.transitionTo(CircuitClosed)
		}
	case CircuitClosed:
		// Already closed, nothing to do
	}
}

// RecordFailure records a failed request.
// May open the circuit if thresholds are exceeded.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalRequests++
	cb.totalFailures++
	cb.consecutiveFailures++
	cb.consecutiveSuccesses = 0
	cb.lastFailure = time.Now()

	switch cb.state {
	case CircuitClosed:
		// Check consecutive failure threshold
		if cb.consecutiveFailures >= cb.config.FailureThreshold {
			cb.transitionTo(CircuitOpen)
			return
		}

		// Check failure rate threshold
		if cb.totalRequests >= cb.config.MinimumRequests {
			failureRate := float64(cb.totalFailures) / float64(cb.totalRequests)
			if failureRate >= cb.config.FailureRateThreshold {
				cb.transitionTo(CircuitOpen)
			}
		}

	case CircuitHalfOpen:
		// Any failure in half-open state reopens the circuit
		cb.transitionTo(CircuitOpen)
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state with cleared counters
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitClosed
	cb.consecutiveFailures = 0
	cb.consecutiveSuccesses = 0
	cb.totalRequests = 0
	cb.totalFailures = 0
	cb.lastStateChange = time.Now()
}

// Stats returns current statistics for monitoring
func (cb *CircuitBreaker) Stats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	var failureRate float64
	if cb.totalRequests > 0 {
		failureRate = float64(cb.totalFailures) / float64(cb.totalRequests)
	}

	return CircuitBreakerStats{
		State:                cb.state,
		ConsecutiveFailures:  cb.consecutiveFailures,
		ConsecutiveSuccesses: cb.consecutiveSuccesses,
		TotalRequests:        cb.totalRequests,
		TotalFailures:        cb.totalFailures,
		FailureRate:          failureRate,
		LastFailure:          cb.lastFailure,
		LastStateChange:      cb.lastStateChange,
	}
}

// transitionTo changes the circuit state (must be called with lock held)
func (cb *CircuitBreaker) transitionTo(newState CircuitState) {
	if cb.state == newState {
		return
	}

	cb.state = newState
	cb.lastStateChange = time.Now()

	// Reset counters on state change
	switch newState {
	case CircuitClosed:
		cb.consecutiveFailures = 0
		cb.consecutiveSuccesses = 0
		cb.totalRequests = 0
		cb.totalFailures = 0
	case CircuitHalfOpen:
		cb.consecutiveSuccesses = 0
	case CircuitOpen:
		cb.consecutiveSuccesses = 0
	}
}

// CircuitBreakerStats contains statistics about the circuit breaker
type CircuitBreakerStats struct {
	State                CircuitState
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
	TotalRequests        int
	TotalFailures        int
	FailureRate          float64
	LastFailure          time.Time
	LastStateChange      time.Time
}

// CircuitOpenError is returned when a request is rejected due to open circuit
type CircuitOpenError struct {
	Provider    string
	State       CircuitState
	LastFailure time.Time
	RetryAfter  time.Duration
}

func (e *CircuitOpenError) Error() string {
	return "circuit breaker is open for provider " + e.Provider + "; retry after " + e.RetryAfter.String()
}
