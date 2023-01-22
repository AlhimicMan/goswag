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
	Name         string `json:"Name"`
	private      int
	NonJSONField uint64 `json:"-"`
}

type SimpleWithPtr struct {
	Name *string `json:"Name"`
}
type WithSubStruct struct {
	Name string       `json:"Name"`
	Sub  SimpleStruct `json:"sub_val"`
}

type WithSubPtrStruct struct {
	Name string        `json:"Name"`
	Sub  *SimpleStruct `json:"sub_val"`
}

type WithUUID struct {
	ID uuid.UUID `json:"id"`
}
type EventTime struct {
	time.Time
}

type EventFinishTime struct {
	time.Time
}

type SpecContainer [2]string
type RecNumber int8
type WithContainers struct {
	Values         []int64            `json:"values"`
	Mapping        map[string]float64 `json:"mapping"`
	Container      SpecContainer      `json:"container"`
	RecNums        []RecNumber        `json:"numbers"`
	SessionTimes   []time.Time        `json:"session_times"`
	EvtTimes       []EventTime        `json:"evt_times"`
	EvtFinishTimes []*EventFinishTime `json:"evt_finish_times"`
}
type UUIDAlias uuid.UUID // Bad idea to make such aliases
type UserName string

type WithTypeAliases struct {
	ID       UUIDAlias  `json:"id"`
	Name     UserName   `json:"Name"`
	Started  time.Time  `json:"started"`
	Finished *time.Time `json:"finished"`
}

type UserID struct {
	uuid.UUID
}
type GroupID struct {
	uuid.UUID
}

type WithEmbeddedTypes struct {
	ID       UserID           `json:"id"`
	GroupID  *GroupID         `json:"alter_id"`
	Targets  []UserID         `json:"targets"`
	Started  EventTime        `json:"started"`
	Finished *EventFinishTime `json:"finished"`
}

type UserNum uint64
type WithInterface struct {
	Number UserNum     `json:"number"`
	Value  interface{} `json:"value"`
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
					if pName == "Name" {
						assert.Equal(t, "string", prop.Type[0])
					} else if pName == "sub_val" {
						refStr := prop.Ref.String()
						assert.True(t, strings.Contains(refStr, "SimpleStruct"))
					} else {
						t.Errorf("struct WithSubStruct unexpected field Name %s", pName)
					}

				}
			} else if strings.Contains(name, "SimpleStruct") {
				assert.Equal(t, 1, len(schema.Properties))
			} else {
				t.Errorf("unexpected definition Name %s", name)
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
	structDef, ok := defs[definitionPrefix+"generator.WithUUID"]
	if !ok {
		t.Fatalf("not found expected definition for WithContainers")
	}
	assert.Equal(t, "object", structDef.Type[0])
	assert.Equal(t, 1, len(structDef.Properties))
	structProp, ok := structDef.Properties["id"]
	if !ok {
		t.Errorf("not found prop for id field")
	} else {
		assert.Equal(t, "string", structProp.Type[0])
		assert.Equal(t, "uuid", structProp.Format)
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
		t.Errorf("not found prop for alias field with alias type")
	} else {
		assert.Equal(t, "array", aliasSlice.Type[0])
		aSliceItems := aliasSlice.Items
		assert.NotNil(t, aSliceItems)
		assert.Equal(t, "integer", aSliceItems.Schema.Type[0])
		assert.Equal(t, "int8", aSliceItems.Schema.Format)
	}

	sessionTimes, ok := structDef.Properties["session_times"]
	if !ok {
		t.Errorf("not found prop for field with list of time")
	} else {
		assert.Equal(t, "array", aliasSlice.Type[0])
		items := sessionTimes.Items
		assert.NotNil(t, items)
		assert.Equal(t, "string", items.Schema.Type[0])
		assert.Equal(t, "date-time", items.Schema.Format)
	}

	evtTimes, ok := structDef.Properties["evt_times"]
	if !ok {
		t.Errorf("not found prop for field with list of embedded time.Time")
	} else {
		assert.Equal(t, "array", evtTimes.Type[0])
		items := evtTimes.Items
		assert.NotNil(t, items)
		assert.Equal(t, "string", items.Schema.Type[0])
		assert.Equal(t, "date-time", items.Schema.Format)
	}

	aliasTimeSlicePointers, ok := structDef.Properties["evt_finish_times"]
	if !ok {
		t.Errorf("not found prop for field with list of pointers to embedded time.Time")
	} else {
		assert.Equal(t, "array", aliasTimeSlicePointers.Type[0])
		items := aliasTimeSlicePointers.Items
		assert.NotNil(t, items)
		assert.Equal(t, "string", items.Schema.Type[0])
		assert.Equal(t, "date-time", items.Schema.Format)
	}

}

func TestGetSchemaWithAliases(t *testing.T) {
	sVal := WithTypeAliases{}
	sType := reflect.TypeOf(sVal)
	sGenerator := NewSchemaGenerator()
	defs, err := sGenerator.GetSchema(sType)
	if err != nil {
		t.Errorf("cannot process WithTypeAliases: %v", err)
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
	nameDef, ok := schema.Properties["Name"]
	if !ok {
		t.Errorf("field Name not found")
	}
	assert.Equal(t, "string", nameDef.Type[0])

	startedDef, ok := schema.Properties["started"]
	if !ok {
		t.Errorf("field started not found")
	}
	assert.Equal(t, "string", startedDef.Type[0])
	assert.Equal(t, "date-time", startedDef.Format)

	finishedDef, ok := schema.Properties["finished"]
	if !ok {
		t.Errorf("field finished not found")
	}
	assert.Equal(t, "string", finishedDef.Type[0])
	assert.Equal(t, "date-time", finishedDef.Format)
}

func TestGetSchemaWithEmbeddedTypes(t *testing.T) {
	sVal := WithEmbeddedTypes{}
	sType := reflect.TypeOf(sVal)
	sGenerator := NewSchemaGenerator()
	defs, err := sGenerator.GetSchema(sType)
	if err != nil {
		t.Errorf("cannot process WithEmbeddedTypes: %v", err)
	}
	if len(defs) != 1 {
		t.Errorf("for WithTypeAliases want 1 definition, have %d", len(defs))
	}
	schema, ok := defs[definitionPrefix+"generator.WithEmbeddedTypes"]
	if !ok {
		t.Errorf("definition WithEmbeddedTypes not found")
	}
	idDef, ok := schema.Properties["id"]
	if !ok {
		t.Errorf("field id not found")
	} else {

	}
	assert.Equal(t, "string", idDef.Type[0])
	assert.Equal(t, "uuid", idDef.Format)
	alterIDDef, ok := schema.Properties["alter_id"]
	if !ok {
		t.Errorf("field Name not found")
	}
	assert.Equal(t, "string", alterIDDef.Type[0])
	assert.Equal(t, "uuid", alterIDDef.Format)

	uuidSliceDef, ok := schema.Properties["targets"]
	if !ok {
		t.Errorf("field Name not found")
	}
	assert.Equal(t, "array", uuidSliceDef.Type[0])
	uuidSliceItems := uuidSliceDef.Items
	assert.NotNil(t, uuidSliceItems)
	assert.Equal(t, "string", uuidSliceItems.Schema.Type[0])
	assert.Equal(t, "uuid", uuidSliceItems.Schema.Format)

	startedDef, ok := schema.Properties["started"]
	if !ok {
		t.Errorf("field started not found")
	}
	assert.Equal(t, "string", startedDef.Type[0])
	assert.Equal(t, "date-time", startedDef.Format)

	finishedDef, ok := schema.Properties["finished"]
	if !ok {
		t.Errorf("field finished not found")
	}
	assert.Equal(t, "string", finishedDef.Type[0])
	assert.Equal(t, "date-time", finishedDef.Format)
}

func TestGetSchemaWithInterface(t *testing.T) {
	sVal := WithInterface{}
	sType := reflect.TypeOf(sVal)
	sGenerator := NewSchemaGenerator()
	defs, err := sGenerator.GetSchema(sType)
	if err != nil {
		t.Errorf("cannot process WithInterface: %v", err)
	}
	structDef, ok := defs[definitionPrefix+"generator.WithInterface"]
	if !ok {
		t.Fatalf("not found expected definition for WithInterface")
	}
	assert.Equal(t, "object", structDef.Type[0])
	assert.Equal(t, 2, len(structDef.Properties))
	numProp, ok := structDef.Properties["number"]
	if !ok {
		t.Errorf("not found prop for id field")
	} else {
		assert.Equal(t, "integer", numProp.Type[0])
		assert.Equal(t, "int64", numProp.Format)
		assert.NotNil(t, numProp.Minimum)
		assert.Equal(t, float64(0), *numProp.Minimum)
	}
	interfaceProp, ok := structDef.Properties["value"]
	if !ok {
		t.Errorf("not found prop for value field tepe interface")
	} else {
		assert.Greater(t, 1, len(interfaceProp.Type))
	}
}
