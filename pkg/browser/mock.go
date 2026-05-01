package browser

import "context"

// Call records a single method invocation on the mock.
type Call struct {
	Method string
	Args   []string
}

// Mock implements Browser by recording calls for test assertions.
type Mock struct {
	Calls  []Call
	Errors map[string]error
}

// NewMock creates a Mock with no preconfigured errors.
func NewMock() *Mock {
	return &Mock{
		Errors: make(map[string]error),
	}
}

func (m *Mock) record(method string, args ...string) error {
	m.Calls = append(m.Calls, Call{Method: method, Args: args})
	if err, ok := m.Errors[method]; ok {
		return err
	}
	return nil
}

func (m *Mock) Navigate(_ context.Context, url string) error {
	return m.record("Navigate", url)
}

func (m *Mock) WaitVisible(_ context.Context, sel string) error {
	return m.record("WaitVisible", sel)
}

func (m *Mock) Click(_ context.Context, sel string) error {
	return m.record("Click", sel)
}

func (m *Mock) SendKeys(_ context.Context, sel string, value string) error {
	return m.record("SendKeys", sel, value)
}

// CallCount returns the number of times a method was called.
func (m *Mock) CallCount(method string) int {
	count := 0
	for _, c := range m.Calls {
		if c.Method == method {
			count++
		}
	}
	return count
}

// CallsTo returns all calls to a specific method.
func (m *Mock) CallsTo(method string) []Call {
	var result []Call
	for _, c := range m.Calls {
		if c.Method == method {
			result = append(result, c)
		}
	}
	return result
}

// Reset clears all recorded calls.
func (m *Mock) Reset() {
	m.Calls = nil
}

