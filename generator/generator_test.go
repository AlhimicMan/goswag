package generator

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
	"time"
)

type SimpleStruct struct {
	Name         string `json:"name"`
	private      int
	NonJSONField uint64 `json:"-"`
}

type SimpleWithPtr struct {
	Name *string `json:"name"`
}
type WithSubStruct struct {
	Name string       `json:"name"`
	Sub  SimpleStruct `json:"sub_val"`
}

type WithSubPtrStruct struct {
	Name string        `json:"name"`
	Sub  *SimpleStruct `json:"sub_val"`
}

type WithUUID struct {
	ID uuid.UUID `json:"id"`
}

type SpecContainer [2]string
type RecNumber int8
type WithContainers struct {
	Values    []int64            `json:"values"`
	Mapping   map[string]float64 `json:"mapping"`
	Container SpecContainer      `json:"container"`
	RecNums   []RecNumber        `json:"numbers"`
}
type UserID uuid.UUID
type UserName string

type WithTypeAliases struct {
	ID       UserID     `json:"id"`
	Name     UserName   `json:"name"`
	Started  time.Time  `json:"started"`
	Finished *time.Time `json:"finished"`
}

func TestGetSchemaSimple(t *testing.T) {
	sVal := SimpleStruct{}
	sValWitPth := SimpleWithPtr{}
	testTypes := map[string]reflect.Type{
		"SimpleStruct":  reflect.TypeOf(sVal),
		"SimpleWithPtr": reflect.TypeOf(sValWitPth),
	}
	for sName, sType := range testTypes {
		sGenerator := NewSchemaGenerator()
		defs, err := sGenerator.GetSchema(sType)
		if err != nil {
			t.Errorf("cannot process %s: %v", sName, err)
		}
		if len(defs) != 1 {
			t.Errorf("for %s want 1 definition, have %d", sName, len(defs))
		}
		for _, schema := range defs {
			assert.Equal(t, schema.Type[0], "object")
			assert.Equal(t, len(schema.Properties), 1)
			for _, prop := range schema.Properties {
				assert.Equal(t, prop.Type[0], "string")
			}
		}
	}
}

func TestGetSchemaWithSubStruct(t *testing.T) {
	sVal := WithSubStruct{}
	sValWitPth := WithSubPtrStruct{}
	testTypes := map[string]reflect.Type{
		"WithSubStruct":    reflect.TypeOf(sVal),
		"WithSubPtrStruct": reflect.TypeOf(sValWitPth),
	}
	for sName, sType := range testTypes {
		sGenerator := NewSchemaGenerator()
		t.Logf("proccessing %s", sType.Name())
		defs, err := sGenerator.GetSchema(sType)
		if err != nil {
			t.Errorf("cannot process %s: %v", sName, err)
		}
		assert.Equal(t, 2, len(defs))
		for name, schema := range defs {
			assert.Equal(t, "object", schema.Type[0])

			if strings.Contains(name, sName) {
				assert.Equal(t, 2, len(schema.Properties))
				for pName, prop := range schema.Properties {
					if pName == "name" {
						assert.Equal(t, "string", prop.Type[0])
					} else if pName == "sub_val" {
						refStr := prop.Ref.String()
						assert.True(t, strings.Contains(refStr, "SimpleStruct"))
					} else {
						t.Errorf("struct WithSubStruct unexpected field name %s", pName)
					}

				}
			} else if strings.Contains(name, "SimpleStruct") {
				assert.Equal(t, 1, len(schema.Properties))
			} else {
				t.Errorf("unexpected definition name %s", name)
			}

		}
	}
}

func TestGetSchemaWithUUID(t *testing.T) {
	sVal := WithUUID{}
	sType := reflect.TypeOf(sVal)
	sGenerator := NewSchemaGenerator()
	defs, err := sGenerator.GetSchema(sType)
	if err != nil {
		t.Errorf("cannot process WithUUID: %v", err)
	}
	if len(defs) != 1 {
		t.Errorf("for WithUUID want 1 definition, have %d", len(defs))
	}
	for _, schema := range defs {
		assert.Equal(t, "object", schema.Type[0])
		assert.Equal(t, 1, len(schema.Properties))
		for _, prop := range schema.Properties {
			assert.Equal(t, "string", prop.Type[0])
			assert.Equal(t, "uuid", prop.Format)
		}
	}
}

func TestGetSchemaWithContainers(t *testing.T) {
	sVal := WithContainers{}
	sType := reflect.TypeOf(sVal)
	sGenerator := NewSchemaGenerator()
	defs, err := sGenerator.GetSchema(sType)
	if err != nil {
		t.Errorf("cannot process WithContainers: %v", err)
	}
	if len(defs) != 1 {
		t.Errorf("for WithUUID want 1 definition, have %d", len(defs))
	}
	structDef, ok := defs[definitionPrefix+"generator.WithContainers"]
	if !ok {
		t.Fatalf("not found expected definition for WithContainers")
	}
	sliceProp, ok := structDef.Properties["values"]
	if !ok {
		t.Errorf("not found prop for array field")
	} else {
		assert.Equal(t, "array", sliceProp.Type[0])
		sliceItems := sliceProp.Items
		assert.NotNil(t, sliceItems)
		assert.Equal(t, "integer", sliceItems.Schema.Type[0])
		assert.Equal(t, "int64", sliceItems.Schema.Format)
	}

	mapProp, ok := structDef.Properties["mapping"]
	if !ok {
		t.Errorf("not found prop for map field")
	} else {
		assert.Equal(t, "object", mapProp.Type[0])
	}

	arrayProp, ok := structDef.Properties["container"]
	if !ok {
		t.Errorf("not found prop for alias type array field")
	} else {
		assert.Equal(t, "array", arrayProp.Type[0])
		arrayItems := arrayProp.Items
		assert.NotNil(t, arrayItems)
		assert.Equal(t, "string", arrayItems.Schema.Type[0])
	}

	aliasSlice, ok := structDef.Properties["numbers"]
	if !ok {
		t.Errorf("not found prop for alice field with alias type")
	} else {
		assert.Equal(t, "array", aliasSlice.Type[0])
		aSliceItems := aliasSlice.Items
		assert.NotNil(t, aSliceItems)
		assert.Equal(t, "integer", aSliceItems.Schema.Type[0])
		assert.Equal(t, "int8", aSliceItems.Schema.Format)
	}
}

func TestGetSchemaWithAliases(t *testing.T) {
	sVal := WithTypeAliases{}
	sType := reflect.TypeOf(sVal)
	sGenerator := NewSchemaGenerator()
	defs, err := sGenerator.GetSchema(sType)
	if err != nil {
		t.Errorf("cannot process WithUUID: %v", err)
	}
	if len(defs) != 1 {
		t.Errorf("for WithTypeAliases want 1 definition, have %d", len(defs))
	}
	schema, ok := defs[definitionPrefix+"generator.WithTypeAliases"]
	if !ok {
		t.Errorf("definition WithTypeAliases not found")
	}
	idDef, ok := schema.Properties["id"]
	if !ok {
		t.Errorf("field id not found")
	}
	// With type alias cannot detect uuid
	assert.Equal(t, "array", idDef.Type[0])
	assert.Equal(t, "", idDef.Format)
	nameDef, ok := schema.Properties["name"]
	if !ok {
		t.Errorf("field name not found")
	}
	assert.Equal(t, "string", nameDef.Type[0])

}
