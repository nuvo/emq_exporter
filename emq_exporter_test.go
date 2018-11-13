package main

import (
	"testing"
)

func Test_parseString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			"easy float parsing",
			args{s: "0.5"},
			0.5,
			false,
		},
		{
			"parse byte represented as string",
			args{s: "123.19M"},
			1.29174077e+08,
			false,
		},
		{
			"invalid string",
			args{s: "invalid string"},
			0,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseString(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseString() = %v, want %v", got, tt.want)
			}
		})
	}
}
