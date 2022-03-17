package cfg2env

// Option represents configuration for Exporter
type Option func(e *Exporter)

// WithEnvironmentTagName sets custom tag name to determine environment variable name
// Default: envconfig
func WithEnvironmentTagName(v string) Option {
	return func(e *Exporter) {
		e.environmentTagName = v
	}
}

// WithDefaultValueTagName sets custom name to determine default value tag name
// Default: default
func WithDefaultValueTagName(v string) Option {
	return func(e *Exporter) {
		e.defaultValueTagName = v
	}
}

// WithExportedFileName sets custom name for resulting file
// Can be relative or absolute path
// File will be overwritten or created
// Default: .env
func WithExportedFileName(v string) Option {
	return func(e *Exporter) {
		e.fileName = v
	}
}

// WithExcludedFields excludes fields from being parsed
// Field can be struct as well (composite literal)
// Default: [RWMutex]
func WithExcludedFields(fields ...string) Option {
	return func(e *Exporter) {
		for _, f := range fields {
			e.excludedFields = append(e.excludedFields, f)
		}
	}
}

// WithHeaderText changes default header for generated file
// Can be set to empty string to disable header
// Default: `# Default configuration`
func WithHeaderText(t string) Option {
	return func(e *Exporter) {
		e.headerText = t
	}
}
