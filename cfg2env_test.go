package cfg2env

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type simpleStruct struct {
	A string `env:"A" default:"a"`
	B int    `env:"B" default:"42"`
}

type nestedStruct struct {
	Outer string `env:"OUTER" default:"outer"`
	Inner struct {
		X string `env:"X" default:"x"`
		Y int    `env:"Y" default:"7"`
	}
}

type tagStruct struct {
	A string `env:"A" default:"a" desc:"descA" validate:"oneof=foo bar"`
	B string `env:"B" default:"b"`
}

type excludedStruct struct {
	Keep    string `env:"KEEP" default:"keep"`
	RWMutex int    // should be excluded by default
}

type allTypes struct {
	S  string        `env:"S"  default:"s"`
	I  int           `env:"I"  default:"1"`
	F  float64       `env:"F"  default:"1.1"`
	B  bool          `env:"B"  default:"true"`
	SS []string      `env:"SS" default:"a,b"`
	II []int         `env:"II" default:"1,2"`
	D  time.Duration `env:"D"  default:"5s"`
}

func TestReflectCfg_SimpleStruct(t *testing.T) {
	e := New(WithEnvironmentTagName("env"), WithDefaultValueTagName("default"))
	items := e.reflectCfg(&simpleStruct{}, "")

	assert.Len(t, items, 4)
	assert.Equal(t, "A", items[1].envVarName)
	assert.Equal(t, "a", items[1].defValue)
	assert.Equal(t, "B", items[3].envVarName)
	assert.Equal(t, "42", items[3].defValue)
}

func TestReflectCfg_NestedStruct(t *testing.T) {
	e := New(WithEnvironmentTagName("env"), WithDefaultValueTagName("default"))
	items := e.reflectCfg(&nestedStruct{}, "")

	// Should include group for Inner and its fields
	var foundInner, foundX, foundY bool
	for _, item := range items {
		if item.nestedGroup && item.comment == "Inner" {
			foundInner = true
		}
		if item.envVarName == "X" && item.defValue == "x" {
			foundX = true
		}
		if item.envVarName == "Y" && item.defValue == "7" {
			foundY = true
		}
	}
	assert.True(t, foundInner)
	assert.True(t, foundX)
	assert.True(t, foundY)
}

func TestReflectCfg_TagExtraction(t *testing.T) {
	e := New(
		WithEnvironmentTagName("env"),
		WithDefaultValueTagName("default"),
		WithDescriptionTagName("desc"),
		WithExtraTagExtraction("validate"),
	)
	items := e.reflectCfg(&tagStruct{}, "")

	// Should include description and validate tag as comments
	var foundDesc, foundValidate bool
	for _, item := range items {
		if item.comment == "A (string) descA" {
			foundDesc = true
		}
		if item.comment == "Tag: validate -> oneof=foo bar" {
			foundValidate = true
		}
	}
	assert.True(t, foundDesc)
	assert.True(t, foundValidate)
}

func TestReflectCfg_ExcludedFields(t *testing.T) {
	e := New(WithEnvironmentTagName("env"), WithDefaultValueTagName("default"))
	items := e.reflectCfg(&excludedStruct{}, "")

	// Should not include RWMutex
	for _, item := range items {
		assert.NotEqual(t, "RWMutex", item.envVarName)
	}
}

func TestReflectCfg_AllTypes(t *testing.T) {
	e := New(WithEnvironmentTagName("env"), WithDefaultValueTagName("default"))
	items := e.reflectCfg(&allTypes{}, "")

	expected := map[string]string{
		"S":  "s",
		"I":  "1",
		"F":  "1.1",
		"B":  "true",
		"SS": "a,b",
		"II": "1,2",
		"D":  "5s",
	}
	for _, item := range items {
		if item.envVarName != "" {
			assert.Equal(t, expected[item.envVarName], item.defValue)
		}
	}
}

func TestReflectCfg_UnexportedFields(t *testing.T) {
	type s struct {
		Exported   string `env:"EXPORTED"   default:"ok"`
		unexported string `env:"UNEXPORTED" default:"fail"` //nolint:unused
	}
	e := New(WithEnvironmentTagName("env"), WithDefaultValueTagName("default"))
	items := e.reflectCfg(&s{}, "")

	for _, item := range items {
		assert.NotEqual(t, "UNEXPORTED", item.envVarName)
	}
}

func TestReflectCfg_EmptyStruct(t *testing.T) {
	e := New()
	items := e.reflectCfg(&struct{}{}, "")
	assert.Len(t, items, 0)
}

func TestReflectCfg_NilPointer(t *testing.T) {
	e := New()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should panic on nil pointer")
		}
	}()
	var s *simpleStruct
	_ = e.reflectCfg(s, "")
}

func TestReflectCfg_Prefix(t *testing.T) {
	type inner struct {
		X string `env:"X" default:"x"`
	}
	type outer struct {
		Inner inner
	}
	e := New(WithEnvironmentTagName("env"), WithDefaultValueTagName("default"))
	items := e.reflectCfg(&outer{}, "prefix.")

	// Should include group with comment "prefix.Inner"
	var found bool
	for _, item := range items {
		if item.nestedGroup && item.comment == "prefix.Inner" {
			found = true
		}
	}
	assert.True(t, found)
}

func TestExportEmptyStruct(t *testing.T) {
	e := New(
		WithEnvironmentTagName("env"),
		WithDefaultValueTagName("default"),
		WithHeaderText("# Loriem ipsum dolor sit amet."),
		WithExcludedFields("TestExcluded"),
		WithExtraEntry("COMPOSE_PROJECT_NAME", "cfg2env"),
		WithExtraTagExtraction("validate"),
		WithExportedFileName(".env.empty"),
	)

	// -- file
	if err := e.ToFile(&struct{}{}); err != nil {
		t.Error(err)
	}

	// -- data
	d, err := e.Export(&struct{}{})
	if err != nil {
		t.Error(err)
	}

	expected := "# Loriem ipsum dolor sit amet.\n\n# Extra pre-declared entries\nCOMPOSE_PROJECT_NAME=cfg2env\n"
	assert.Equal(t, expected, string(d))
}

type testConfig struct {
	sync.RWMutex

	A string `env:"A" default:"def_value_of_a"
        desc:"Just a dummy value for purpose of this test
        and should not be used as real example, this text is 
        just here for placeholder ... testing testing"`
	B            string `env:"B" default:"def_value_of_b"`
	C            string `env:"C" default:"def_value_of_c" validate:"oneof=one two three"`
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
	DeepNested struct {
		DeepNested2 struct {
			DeepNested3 struct {
				Foo string `env:"DEEP_NESTED_FOO" default:"foo"`
				Bar string `env:"DEEP_NESTED_BAR" default:"bar"`
			}
		}
	}
}

func TestExport(t *testing.T) {
	e := New(
		WithEnvironmentTagName("env"),
		WithDefaultValueTagName("default"),
		WithHeaderText("# Test Header"),
		WithExcludedFields("TestExcluded"),
		WithExtraEntry("COMPOSE_PROJECT_NAME", "cfg2env"),
		WithExtraTagExtraction("validate"),
	)

	// -- file
	if err := e.ToFile(new(testConfig)); err != nil {
		t.Error(err)
	}

	// -- data
	d, err := e.Export(new(testConfig))
	if err != nil {
		t.Error(err)
	}

	expected := "# Test Header\n\n# Extra pre-declared entries\nCOMPOSE_PROJECT_NAME=cfg2env\n\n# A (string) Just a dummy value for purpose of this test\n# and should not be used as real example, this text is \n# just here for placeholder ... testing testing\nA=def_value_of_a\n# B (string)\nB=def_value_of_b\n# C (string)\n# Tag: validate -> oneof=one two three\nC=def_value_of_c\n\n## Nested\n\n# Foo (int8) Simple dummy value for testing\nNESTED_FOO=98\n# Bar ([]string) Simple dummy value for testing\nNESTED_BAR=one,two,three\n\n## Nested.NestedTwo\n\n# Foo ([]int64) Simple dummy value for testing\nNESTED_NESTED2_FOO=1,2,3,4,5,6,7,8,9,0\n# Bar (time.Duration) Simple dummy value for testing\nNESTED_NESTED2_BAR=10s\n\n## DeepNested\n\n## DeepNested.DeepNested2\n\n## DeepNested.DeepNested2.DeepNested3\n\n# Foo (string)\nDEEP_NESTED_FOO=foo\n# Bar (string)\nDEEP_NESTED_BAR=bar\n"
	assert.Equal(t, expected, string(d))
}

type testConfigNestedFirst struct {
	NestedFirst struct {
		A string `env:"NestedFirst_A" default:"def_value_of_a"`
		B string `env:"NestedFirst_B" default:"def_value_of_b"`
		C string `env:"NestedFirst_C" default:"def_value_of_c"`
	}
	A string `env:"A" default:"def_value_of_a"`
	B string `env:"B" default:"def_value_of_b"`
	C string `env:"C" default:"def_value_of_c"`
}

func TestExportNestedFirst(t *testing.T) {
	t.Skip()

	e := New(
		WithEnvironmentTagName("env"),
		WithDefaultValueTagName("default"),
		WithHeaderText("# Test Header"),
		WithExcludedFields("TestExcluded"),
		WithExtraEntry("COMPOSE_PROJECT_NAME", "cfg2env"),
		WithExtraTagExtraction("validate"),
		WithExportedFileName(".env.nested_first"),
	)

	// -- file
	if err := e.ToFile(new(testConfigNestedFirst)); err != nil {
		t.Error(err)
	}

	// -- data
	d, err := e.Export(new(testConfigNestedFirst))
	if err != nil {
		t.Error(err)
	}

	expected := "# Test Header\n\n# Extra pre-declared entries\nCOMPOSE_PROJECT_NAME=cfg2env\n\n## NestedFirst\n\n# A (string)\nNestedFirst_A=def_value_of_a\n# B (string)\nNestedFirst_B=def_value_of_b\n# C (string)\nNestedFirst_C=def_value_of_c\n\n# A (string)\nA=def_value_of_a\n# B (string)\nB=def_value_of_b\n# C (string)\nC=def_value_of_c\n"
	assert.Equal(t, expected, string(d))
}
