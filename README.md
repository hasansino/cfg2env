[![Go Report Card](https://goreportcard.com/badge/github.com/hasansino/cfg2env)](https://goreportcard.com/report/github.com/hasansino/cfg2env)
[![Build Status](https://travis-ci.com/hasansino/cfg2env.svg?branch=master)](https://travis-ci.com/hasansino/cfg2env)

# cfg2env

cfg2env is a tool to convert configuration objects to .env format.  

Features:
* Text header for generated files
* Configurable tag names (environment variable name, default value and description)
* Excluded fields
* (Optional) Description tags with multi-line support (see example)

## Installation

```bash
~ $ go get -u github.com/hasansino/cfg2env
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

# A (string) Just a dummy value for purpose of this test
# and should not be used as real example, this text is 
# just here for placeholder ... testing testing
A=def_value_of_a
B=def_value_of_b

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

