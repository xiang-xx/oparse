package core

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func Test_decodePath(t *testing.T) {
	paths := []common.Address{
		common.HexToAddress("0xff970a61a04b1ca14834a43f5de4533ebddb5cc8"),
		common.HexToAddress("0xfd086bc7cd5c481dcc9c85ebe478a1c0b69fcbb9"),
		common.HexToAddress("0x0000000000000000000000000000000000000000"),
	}
	fees := []int{
		500, 3000,
	}
	pathBytes, _ := encodePath(paths, fees)

	type args struct {
		pathByte []byte
	}
	tests := []struct {
		name  string
		args  args
		want  []common.Address
		want1 []int
	}{
		{
			name: "decode path",
			args: args{
				pathByte: pathBytes,
			},
			want:  paths,
			want1: fees,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := decodePath(tt.args.pathByte)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodePath() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("decodePath() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
