package cfg2env

// Option represents configuration for Exporter
type Option func(e *Exporter)

// WithEnvironmentTagName sets custom tag name to determine environment variable name
// Default: env
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

// WithDescriptionTagName sets custom name to determine description tag name
// Default: desc
func WithDescriptionTagName(v string) Option {
	return func(e *Exporter) {
		e.descriptionTagName = v
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
		e.excludedFields = append(e.excludedFields, fields...)
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

// WithExtraEntry adds extra static entry to top of resulting .env file
func WithExtraEntry(key string, value interface{}) Option {
	return func(e *Exporter) {
		e.extraEntries[key] = value
	}
}

// WithExtraTagExtraction adds tag to be included in variable description
func WithExtraTagExtraction(tag string) Option {
	return func(e *Exporter) {
		e.extraTags = append(e.extraTags, tag)
	}
}
