package cfg2env

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
)

// default values
const (
	_defHeaderText          = `# Default configuration`
	_defEnvironmentTagName  = `envconfig`
	_defDefaultValueTagName = `default`
	_defDescriptionTagName  = `desc`
	_defFileName            = `.env`
)

var (
	_defExcludedFields = []string{
		"RWMutex",
	}
)

// Exporter is configuration file parser and exporter
type Exporter struct {
	headerText          string
	environmentTagName  string
	defaultValueTagName string
	descriptionTagName  string
	fileName            string
	excludedFields      []string
	extraEntries        map[string]interface{}
}

// cfgItem represents single item of configuration
// it can be final value (string,int,float etc.) or struct
// which will be recursively parsed
type cfgItem struct {
	nestedGroup bool
	comment     string
	envVarName  string
	defValue    string
}

// New creates new exporter with provided options
func New(opts ...Option) *Exporter {
	e := &Exporter{
		headerText:     _defHeaderText,
		excludedFields: make([]string, 0),
		extraEntries:   make(map[string]interface{}),
	}
	for _, v := range _defExcludedFields {
		e.excludedFields = append(e.excludedFields, v)
	}
	for _, o := range opts {
		o(e) // apply all options if needed
	}
	e.defaults()
	return e
}

func (e *Exporter) defaults() {
	if len(e.environmentTagName) == 0 {
		e.environmentTagName = _defEnvironmentTagName
	}
	if len(e.defaultValueTagName) == 0 {
		e.defaultValueTagName = _defDefaultValueTagName
	}
	if len(e.descriptionTagName) == 0 {
		e.descriptionTagName = _defDescriptionTagName
	}
	if len(e.fileName) == 0 {
		e.fileName = _defFileName
	}
}

// ToFile exports data to file, file path can be set with WithExportedFileName
func (e *Exporter) ToFile(cfg interface{}) error {
	f, err := os.Create(e.fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	data, err := e.Export(cfg)
	if err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	if err := f.Sync(); err != nil {
		log.Fatalf("failed to sync file: %v", err)
	}

	return nil
}

// Export exports struct to readable and structured .env format
func (e *Exporter) Export(cfg interface{}) ([]byte, error) {
	buff := new(bytes.Buffer)
	exported := e.reflectCfg(cfg, ``)

	if len(e.headerText) > 0 {
		if _, err := buff.WriteString(e.headerText + "\n\n"); err != nil {
			return nil, fmt.Errorf("failed to write to file: %v", err)
		}
	}

	if len(e.extraEntries) > 0 {
		if _, err := buff.WriteString("# Extra pre-declared entries\n"); err != nil {
			return nil, fmt.Errorf("failed to write to buffer: %v", err)
		}
		for k, v := range e.extraEntries {
			if _, err := buff.WriteString(
				fmt.Sprintf("%s=%v\n", k, v),
			); err != nil {
				return nil, fmt.Errorf("failed to write to buffer: %v", err)
			}
		}
		if _, err := buff.WriteString("\n"); err != nil {
			return nil, fmt.Errorf("failed to write to buffer: %v", err)
		}
	}

	for i := range exported {
		if len(exported[i].comment) > 0 { // comment
			toWrite := formatComment(exported[i].comment)
			if exported[i].nestedGroup {
				toWrite = "\n##" + toWrite
			}
			toWrite += "\n"
			if exported[i].nestedGroup {
				toWrite += "\n"
			}
			_, err := buff.WriteString(toWrite)
			if err != nil {
				return nil, fmt.Errorf("failed to write to buffer: %v", err)
			}
		} else { // variable=value
			value := exported[i].defValue
			if strings.Contains(exported[i].defValue, " ") {
				value = fmt.Sprintf("\"%s\"", value)
			}
			_, err := buff.WriteString(fmt.Sprintf("%s=%s\n",
				exported[i].envVarName, exported[i].defValue))
			if err != nil {
				return nil, fmt.Errorf("failed to write to buffer: %v", err)
			}
		}
	}

	return buff.Bytes(), nil
}

// reflectCfg exports struct data in structured format
func (e *Exporter) reflectCfg(cfg interface{}, prefix string) []cfgItem {
	var (
		rt = reflect.TypeOf(cfg)
		rv = reflect.ValueOf(cfg)
	)

	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	exported := make([]cfgItem, 0)

	for i := 0; i < rt.NumField(); i++ {
		var (
			field     = rt.Field(i)
			value     = rv.FieldByName(field.Name)
			fieldPath = prefix + field.Name
		)

		// skip unexported
		if len(field.PkgPath) != 0 {
			continue
		}

		// skip excluded fields
		var isExcluded bool
		for _, n := range e.excludedFields {
			if strings.ToLower(field.Name) == strings.ToLower(n) {
				isExcluded = true
				break
			}
		}
		if isExcluded {
			continue
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			exported = append(exported, cfgItem{
				nestedGroup: true,
				comment:     fieldPath,
			})
			exported = append(
				exported,
				e.reflectCfg(value.Addr().Interface(), fieldPath+".")...,
			)
		default:
			tag := MultilineStructTag(field.Tag)
			if envVarName := tag.Get(e.environmentTagName); len(envVarName) > 0 {
				// variable description [field_name (type) description]
				itemDescription := cfgItem{
					comment: fmt.Sprintf(
						"%s (%s)", field.Name, field.Type.String(),
					),
				}
				if desc := tag.Get(e.descriptionTagName); len(desc) > 0 {
					itemDescription.comment += " " + desc
				}
				exported = append(exported, itemDescription)

				// variable definition [variable=default_value]
				exported = append(exported, cfgItem{
					envVarName: envVarName, defValue: tag.Get(e.defaultValueTagName),
				})
			}
		}
	}

	return exported
}

func formatComment(s string) string {
	// replace all tabs with spaces
	tabs := regexp.MustCompile(`\t+`)
	s = tabs.ReplaceAllString(s, " ")
	// truncate all repetitive spaces to one
	spaces := regexp.MustCompile(` {2,}`)
	s = spaces.ReplaceAllString(s, " ")
	// put # char in front of every line
	s = "# " + strings.Replace(s, "\n", "\n#", -1)
	return s
}
