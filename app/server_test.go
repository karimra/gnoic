package app

import (
	"testing"
)

func Test_decimalToOctal(t *testing.T) {
	type args struct {
		d uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "0",
			args: args{
				d: 0,
			},
			want: 0,
		},
		{
			name: "1",
			args: args{
				d: 1,
			},
			want: 1,
		}, {
			name: "9",
			args: args{
				d: 9,
			},
			want: 11,
		},
		{
			name: "100",
			args: args{
				d: 100,
			},
			want: 144,
		},
		{
			name: "101",
			args: args{
				d: 101,
			},
			want: 145,
		},
		{
			name: "511",
			args: args{
				d: 511,
			},
			want: 777,
		},
		{
			name: "1000",
			args: args{
				d: 1000,
			},
			want: 1750,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decimalToOctal(tt.args.d); got != tt.want {
				t.Errorf("decimalToOctal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_octalToDecimal(t *testing.T) {
	type args struct {
		d uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "0",
			args: args{
				d: 0,
			},
			want: 0,
		},
		{
			name: "1",
			args: args{
				d: 1,
			},
			want: 1,
		}, {
			name: "11",
			args: args{
				d: 11,
			},
			want: 9,
		},
		{
			name: "144",
			args: args{
				d: 144,
			},
			want: 100,
		},
		{
			name: "145",
			args: args{
				d: 145,
			},
			want: 101,
		},
		{
			name: "777",
			args: args{
				d: 777,
			},
			want: 511,
		},
		{
			name: "1750",
			args: args{
				d: 1750,
			},
			want: 1000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := octalToDecimal(tt.args.d); got != tt.want {
				t.Errorf("octalToDecimal() = %v, want %v", got, tt.want)
			}
		})
	}
}
