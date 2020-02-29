package vindexes

// NOTE(dweitzman): This is a modified version of hash_test.go

import (
	"reflect"
	"strings"
	"testing"

	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/vt/key"
)

func createHashOffset(offset string) Vindex {
	hv, err := CreateVindex("hash_offset", "nn", map[string]string{"offset": offset})
	if err != nil {
		panic(err)
	}
	return hv
}

func TestHashOffsetCreation(t *testing.T) {
	tcases := []struct {
		offset, wantErr string
	}{
		{
			offset:  "",
			wantErr: "hash_offset.NewHashOffset: 'offset' param missing",
		},
		{
			offset:  "weird",
			wantErr: `hash_offset.NewHashOffset: "weird" is not a valid offset`,
		},
		{
			offset:  "1.2",
			wantErr: `hash_offset.NewHashOffset: "1.2" is not a valid offset`,
		},
	}

	for _, tcase := range tcases {
		params := map[string]string{"offset": tcase.offset}
		_, err := CreateVindex("hash_offset", "nn", params)
		if err == nil || err.Error() != tcase.wantErr {
			t.Errorf(`CreateVindex("hash_offset", "nn", %v) had err %v. Want: %v`, params, err, tcase.wantErr)
		}
	}
}

func TestHashOffsetFunc(t *testing.T) {
	tcases := []struct {
		offset   uint64
		in, want []sqltypes.Value
	}{{
		offset: 1,
		in:     []sqltypes.Value{sqltypes.NewUint64(1), sqltypes.NewUint64(100)},
		want:   []sqltypes.Value{sqltypes.NewUint64(2), sqltypes.NewUint64(101)},
	}, {
		offset: 1,
		in:     []sqltypes.Value{sqltypes.NewUint64(1), sqltypes.NULL, sqltypes.NewUint64(100)},
		want:   []sqltypes.Value{sqltypes.NewUint64(2), sqltypes.NULL, sqltypes.NewUint64(101)},
	}, {
		offset: ^uint64(0),
		in:     []sqltypes.Value{sqltypes.NewUint64(1), sqltypes.NULL, sqltypes.NewUint64(100)},
		want:   []sqltypes.Value{sqltypes.NewUint64(0), sqltypes.NULL, sqltypes.NewUint64(99)},
	}, {
		offset: 2,
		in:     []sqltypes.Value{sqltypes.NewUint64(1), sqltypes.NewUint64(100)},
		want:   []sqltypes.Value{sqltypes.NewUint64(3), sqltypes.NewUint64(102)},
	}}

	for _, tcase := range tcases {
		got := applyOffset(tcase.offset, tcase.in)
		if !reflect.DeepEqual(tcase.want, got) {
			t.Errorf("applyOffset(%v, %v) = %v, want %v", tcase.offset, tcase.in, got, tcase.want)
		}
	}
}

func TestHashOffsetCost(t *testing.T) {
	hashOffset := createHashOffset("1")

	if hashOffset.Cost() != 1 {
		t.Errorf("Cost(): %d, want 1", hashOffset.Cost())
	}
}

func TestHashOffsetString(t *testing.T) {
	hashOffset := createHashOffset("1")

	if strings.Compare("nn", hashOffset.String()) != 0 {
		t.Errorf("String(): %s, want hashOffset", hashOffset.String())
	}
}

func TestHashOffsetMap(t *testing.T) {
	hashOffset := createHashOffset("1")

	got, err := hashOffset.(SingleColumn).Map(nil, []sqltypes.Value{
		sqltypes.NewInt64(0),
		sqltypes.NewInt64(1),
		sqltypes.NewInt64(2),
		sqltypes.NULL,
		sqltypes.NewInt64(3),
		sqltypes.NewInt64(4),
		sqltypes.NewInt64(5),
	})
	if err != nil {
		t.Error(err)
	}
	want := []key.Destination{
		key.DestinationKeyspaceID([]byte("\x16k@\xb4J\xbaK\xd6")),
		key.DestinationKeyspaceID([]byte("\x06\xe7\xea\"Βp\x8f")),
		key.DestinationKeyspaceID([]byte("N\xb1\x90ɢ\xfa\x16\x9c")),
		key.DestinationNone{},
		key.DestinationKeyspaceID([]byte("\xd2\xfd\x88g\xd5\r-\xfe")),
		key.DestinationKeyspaceID([]byte("p\xbb\x02<\x81\f\xa8z")),
		key.DestinationKeyspaceID([]byte("\xf0\x98H\n\xc4ľq")),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}
}

func TestHashOffsetMapNegative(t *testing.T) {
	hashOffset := createHashOffset("-1")

	got, err := hashOffset.(SingleColumn).Map(nil, []sqltypes.Value{
		sqltypes.NewInt64(2),
		sqltypes.NewInt64(3),
		sqltypes.NewInt64(4),
		sqltypes.NULL,
		sqltypes.NewInt64(5),
		sqltypes.NewInt64(6),
		sqltypes.NewInt64(7),
	})
	if err != nil {
		t.Error(err)
	}
	want := []key.Destination{
		key.DestinationKeyspaceID([]byte("\x16k@\xb4J\xbaK\xd6")),
		key.DestinationKeyspaceID([]byte("\x06\xe7\xea\"Βp\x8f")),
		key.DestinationKeyspaceID([]byte("N\xb1\x90ɢ\xfa\x16\x9c")),
		key.DestinationNone{},
		key.DestinationKeyspaceID([]byte("\xd2\xfd\x88g\xd5\r-\xfe")),
		key.DestinationKeyspaceID([]byte("p\xbb\x02<\x81\f\xa8z")),
		key.DestinationKeyspaceID([]byte("\xf0\x98H\n\xc4ľq")),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(): %#v, want %+v", got, want)
	}
}

func TestHashOffsetVerify(t *testing.T) {
	hashOffset := createHashOffset("1")

	ids := []sqltypes.Value{sqltypes.NewInt64(0), sqltypes.NewInt64(1)}
	ksids := [][]byte{[]byte("\x16k@\xb4J\xbaK\xd6"), []byte("\x16k@\xb4J\xbaK\xd6")}
	got, err := hashOffset.(SingleColumn).Verify(nil, ids, ksids)
	if err != nil {
		t.Fatal(err)
	}
	want := []bool{true, false}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("binaryMD5.Verify: %v, want %v", got, want)
	}

	// Failure test
	_, err = hashOffset.(SingleColumn).Verify(nil, []sqltypes.Value{sqltypes.NewVarBinary("aa")}, [][]byte{nil})
	wantErr := "hash.Verify: could not parse value: 'aa'"
	if err == nil || err.Error() != wantErr {
		t.Errorf("hashOffset.Verify err: %v, want %s", err, wantErr)
	}
}

func TestHashOffsetReverseMap(t *testing.T) {
	hashOffset := createHashOffset("1")

	got, err := hashOffset.(Reversible).ReverseMap(nil, [][]byte{[]byte("\x16k@\xb4J\xbaK\xd6")})
	if err != nil {
		t.Error(err)
	}
	want := []sqltypes.Value{sqltypes.NewUint64(uint64(0))}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReverseMap(): %v, want %v", got, want)
	}
}

func TestHashOffsetReverseMapNeg(t *testing.T) {
	hashOffset := createHashOffset("1")

	_, err := hashOffset.(Reversible).ReverseMap(nil, [][]byte{[]byte("\x16k@\xb4J\xbaK\xd6\x16k@\xb4J\xbaK\xd6")})
	want := "invalid keyspace id: 166b40b44aba4bd6166b40b44aba4bd6"
	if err.Error() != want {
		t.Error(err)
	}
}
