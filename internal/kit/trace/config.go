package trace

// A Config is an opentelemetry config.
type Config struct {
	Name     string  `json:",optional"`
	Endpoint string  `json:",optional"`
	Sampler  float64 `json:",default=1.0"`
	Batcher  string  `json:",default=otlpgrpc,options=otlpgrpc|otlphttp|file"`
	// OtlpHeaders represents the headers for OTLP gRPC or HTTP transport.
	// OtlpHeaders map[string]string `json:",optional"`
	// // OtlpHttpPath represents the path for OTLP HTTP transport.
	// OtlpHttpPath string `json:",optional"`
	// Disabled indicates whether StartAgent starts the agent.
	Disabled bool `json:",optional"`
}
