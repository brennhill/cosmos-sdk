package table

import (
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeSafeRowGetter(t *testing.T) {
	storeKey := sdk.NewKVStoreKey("test")
	ctx := NewMockContext()
	const prefixKey = 0x2
	store := prefix.NewStore(ctx.KVStore(storeKey), []byte{prefixKey})
	md := testdata.TableModel{
		Id:   "my-id",
		Name: "some name",
	}
	bz, err := md.Marshal()
	require.NoError(t, err)
	store.Set([]byte("my-id"), bz)

	specs := map[string]struct {
		srcRowID     RowID
		srcModelType reflect.Type
		expObj       interface{}
		expErr       *errors.Error
	}{
		"happy path": {
			srcRowID:     []byte("my-id"),
			srcModelType: reflect.TypeOf(testdata.TableModel{}),
			expObj:       md,
		},
		"unknown rowID should return ErrNotFound": {
			srcRowID:     []byte("unknown"),
			srcModelType: reflect.TypeOf(testdata.TableModel{}),
			expErr:       ErrNotFound,
		},
		"wrong type should cause ErrType": {
			srcRowID:     []byte("my-id"),
			srcModelType: reflect.TypeOf(testdata.Cat{}),
			expErr:       ErrType,
		},
		"empty rowID not allowed": {
			srcRowID:     []byte{},
			srcModelType: reflect.TypeOf(testdata.TableModel{}),
			expErr:       ErrArgument,
		},
		"nil rowID not allowed": {
			srcModelType: reflect.TypeOf(testdata.TableModel{}),
			expErr:       ErrArgument,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			interfaceRegistry := types.NewInterfaceRegistry()
			cdc := codec.NewProtoCodec(interfaceRegistry)

			getter := NewTypeSafeRowGetter(prefixKey, spec.srcModelType, cdc)
			var loadedObj testdata.TableModel

			err := getter(ctx.KVStore(storeKey), spec.srcRowID, &loadedObj)
			if spec.expErr != nil {
				require.True(t, spec.expErr.Is(err), err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, spec.expObj, loadedObj)
		})
	}
}