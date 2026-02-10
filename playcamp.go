package playcamp

// Environment represents the target environment for the API.
type Environment string

const (
	// EnvironmentSandbox targets the sandbox API server.
	EnvironmentSandbox Environment = "sandbox"
	// EnvironmentLive targets the live (production) API server.
	EnvironmentLive Environment = "live"
)

// EnvironmentURL returns the base URL for the given environment.
// Returns an empty string for unknown environments.
func EnvironmentURL(env Environment) string {
	switch env {
	case EnvironmentSandbox:
		return "https://sandbox-sdk-api.playcamp.io"
	case EnvironmentLive:
		return "https://sdk-api.playcamp.io"
	default:
		return ""
	}
}

// Int returns a pointer to the given int value.
func Int(v int) *int { return &v }

// String returns a pointer to the given string value.
func String(v string) *string { return &v }

// Bool returns a pointer to the given bool value.
func Bool(v bool) *bool { return &v }

// Float64 returns a pointer to the given float64 value.
func Float64(v float64) *float64 { return &v }
