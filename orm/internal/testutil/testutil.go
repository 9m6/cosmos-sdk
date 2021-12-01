package testutil

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormfield"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/internal/testpb"
)

// TestFieldSpec defines a test field against the testpb.A message.
type TestFieldSpec struct {
	FieldName protoreflect.Name
	Gen       *rapid.Generator
}

var TestFieldSpecs = []TestFieldSpec{
	{
		"u32",
		rapid.Uint32(),
	},
	{
		"u64",
		rapid.Uint64(),
	},
	{
		"str",
		rapid.String().Filter(func(x string) bool {
			// filter out null terminators
			return strings.IndexByte(x, 0) < 0
		}),
	},
	{
		"bz",
		rapid.SliceOfN(rapid.Byte(), 0, 255),
	},
	{
		"i32",
		rapid.Int32(),
	},
	{
		"f32",
		rapid.Uint32(),
	},
	{
		"s32",
		rapid.Int32(),
	},
	{
		"sf32",
		rapid.Int32(),
	},
	{
		"i64",
		rapid.Int64(),
	},
	{
		"f64",
		rapid.Uint64(),
	},
	{
		"s64",
		rapid.Int64(),
	},
	{
		"sf64",
		rapid.Int64(),
	},
	{
		"b",
		rapid.Bool(),
	},
	{
		"ts",
		rapid.ArrayOf(2, rapid.Int64()).Map(func(xs [2]int64) protoreflect.Message {
			return (&timestamppb.Timestamp{
				Seconds: xs[0],
				Nanos:   int32(xs[1]),
			}).ProtoReflect()
		}),
	},
	{
		"dur",
		rapid.ArrayOf(2, rapid.Int64()).Map(func(xs [2]int64) protoreflect.Message {
			return (&durationpb.Duration{
				Seconds: xs[0],
				Nanos:   int32(xs[1]),
			}).ProtoReflect()
		}),
	},
	{
		"e",
		rapid.Int32().Map(func(x int32) protoreflect.EnumNumber {
			return protoreflect.EnumNumber(x)
		}),
	},
}

func MakeTestCodec(fname protoreflect.Name, nonTerminal bool) (ormfield.Codec, error) {
	field := GetTestField(fname)
	if field == nil {
		return nil, fmt.Errorf("can't find field %s", fname)
	}
	return ormfield.GetCodec(field, nonTerminal)
}

func GetTestField(fname protoreflect.Name) protoreflect.FieldDescriptor {
	a := &testpb.A{}
	return a.ProtoReflect().Descriptor().Fields().ByName(fname)
}

type TestKeyCodec struct {
	KeySpecs []TestFieldSpec
	Codec    *ormkv.KeyCodec
}

var TestKeyCodecGen = rapid.Custom(func(t *rapid.T) TestKeyCodec {
	xs := rapid.SliceOfNDistinct(rapid.IntRange(0, len(TestFieldSpecs)-1), 0, 5, func(i int) int { return i }).
		Draw(t, "fieldSpecs").([]int)

	var specs []TestFieldSpec
	var fields []protoreflect.FieldDescriptor

	for _, x := range xs {
		spec := TestFieldSpecs[x]
		specs = append(specs, spec)
		fields = append(fields, GetTestField(spec.FieldName))
	}

	prefix := rapid.SliceOfN(rapid.Byte(), 0, 5).Draw(t, "prefix").([]byte)

	cdc, err := ormkv.NewKeyCodec(prefix, fields)
	if err != nil {
		panic(err)
	}

	return TestKeyCodec{
		Codec:    cdc,
		KeySpecs: specs,
	}
})

func (k TestKeyCodec) Draw(t *rapid.T, id string) []protoreflect.Value {
	n := len(k.KeySpecs)
	keyValues := make([]protoreflect.Value, n)
	for i, k := range k.KeySpecs {
		keyValues[i] = protoreflect.ValueOf(k.Gen.Draw(t, fmt.Sprintf("%s[%d]", id, i)))
	}
	return keyValues
}
