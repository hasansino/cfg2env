package cfg2env

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

func _newExporter() *Exporter {
	return New(
		WithEnvironmentTagName("env"),
		WithDefaultValueTagName("default"),
		WithHeaderText("# Test Header"),
		WithExcludedFields("TestExcluded"),
		WithExtraEntry("COMPOSE_PROJECT_NAME", "cfg2env"),
	)
}

func TestExportToFile(t *testing.T) {
	e := _newExporter()
	if err := e.ToFile(new(testConfig)); err != nil {
		t.Error(err)
	}
}

func TestExport(t *testing.T) {
	e := _newExporter()
	d, err := e.Export(new(testConfig))
	if err != nil {
		t.Error(err)
	}
	expected := "# Test Header\n\n# Extra pre-declared entries\nCOMPOSE_PROJECT_NAME=cfg2env\n\n# A (string) Just a dummy value for purpose of this test\n# and should not be used as real example, this text is \n# just here for placeholder ... testing testing\nA=def_value_of_a\n# B (string)\nB=def_value_of_b\n\n### Nested\n\n# Foo (int8) Simple dummy value for testing\nNESTED_FOO=98\n# Bar ([]string) Simple dummy value for testing\nNESTED_BAR=one,two,three\n\n### Nested.NestedTwo\n\n# Foo ([]int64) Simple dummy value for testing\nNESTED_NESTED2_FOO=1,2,3,4,5,6,7,8,9,0\n# Bar (time.Duration) Simple dummy value for testing\nNESTED_NESTED2_BAR=10s\n"
	assert.Equal(t, expected, string(d))
}
