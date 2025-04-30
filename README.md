# cfg2env

cfg2env is a tool to convert configuration objects to .env format.

Features:

* Text header for generated files
* Configurable tag names (environment variable name, default value and description)
* Excluded fields
* Description tags with multi-line support (see example)
* Add extra tag to be included in field description

## Installation

```bash
go get github.com/hasansino/cfg2env
```

## Usage

cfg2env is intended to be used in your application code and not as binary.  
It is best practice creating a file under cmd/cfg2env/main.go and manually running it  
since the result will depend on latest version of your configuration object.

## Example

```go
package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/hasansino/cfg2env"
)

func main() {
	type testConfig struct {
		sync.RWMutex

		A string `env:"A" default:"def_value_of_a"
                desc:"Just a dummy value for purpose of this test
                and should not be used as real example, this text is 
                just here for placeholder ... testing testing"`
		B            string `env:"B" default:"def_value_of_b"`
		C            string `env:"C" default:"def_value_of_c" validate:"required"`
		TestExcluded struct {
			Foo int64 `env:"ERROR" default:"ERROR"`
			Bar int64 `env:"ERROR" default:"ERROR"`
		}
		Nested struct {
			Foo int8 `env:"NESTED_FOO" default:"98"
                        desc:"Simple dummy value for testing"`
			Bar []string `env:"NESTED_BAR" default:"one,two,three"
                        desc:"Simple dummy value for testing"`
			NestedTwo struct {
				Foo []int64 `env:"NESTED_NESTED2_FOO" default:"1,2,3,4,5,6,7,8,9,0"
                                desc:"Simple dummy value for testing"`
				Bar time.Duration `env:"NESTED_NESTED2_BAR" default:"10s"
                                desc:"Simple dummy value for testing"`
			}
		}
	}

	exporter := cfg2env.New(
		cfg2env.WithEnvironmentTagName("env"),
		cfg2env.WithDefaultValueTagName("default"),
		cfg2env.WithHeaderText("# Test Header"),
		cfg2env.WithExcludedFields("TestExcluded"),
		cfg2env.WithExtraEntry("COMPOSE_PROJECT_NAME", "cfg2env"),
		cfg2env.WithExtraTagExtraction("validate"),
	)

	err := exporter.ToFile(new(testConfig))
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
```

This will generate .env file in project root:

```dotenv
# Test Header

# Extra pre-declared entries
COMPOSE_PROJECT_NAME=cfg2env

# A (string) Just a dummy value for purpose of this test
# and should not be used as real example, this text is 
# just here for placeholder ... testing testing
A=def_value_of_a
# B (string)
B=def_value_of_b
# C (string)
# Tag: validate -> required
C=def_value_of_c

### Nested

# Foo (int8) Simple dummy value for testing
NESTED_FOO=98
# Bar ([]string) Simple dummy value for testing
NESTED_BAR=one,two,three

### Nested.NestedTwo

# Foo ([]int64) Simple dummy value for testing
NESTED_NESTED2_FOO=1,2,3,4,5,6,7,8,9,0
# Bar (time.Duration) Simple dummy value for testing
NESTED_NESTED2_BAR=10s
```
