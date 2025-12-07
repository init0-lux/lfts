package state

// GlobalState is the singleton state instance
var GlobalState = NewStorage()

// Set stores a value in global state
func Set(key string, value []byte) error {
	return GlobalState.Set(key, value)
}

// Get retrieves a value from global state
func Get(key string) ([]byte, error) {
	return GlobalState.Get(key)
}

// Has checks if a key exists in global state
func Has(key string) bool {
	return GlobalState.Has(key)
}

