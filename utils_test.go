package gocli

import (
	"net"
	"reflect"
	"testing"
	"time"
)

func Test_getFlagArgValue(t *testing.T) {
	IP_str := "10.100.10.10"

	tm, _ := time.Parse("01/01/2006 03:04:05 PM", "06/12/2023 03:04:05 PM")

	ip := net.ParseIP(IP_str)

	type args struct {
		fa    IFlagArg
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "String type",
			args: args{
				fa: &Flag[String]{
					Name: "f1",
				},
				value: "aaa",
			},
			want: "aaa",
		},
		{
			name: "[]String type",
			args: args{
				fa: &Flag[[]String]{
					Name: "f1",
				},
				value: "aaa",
			},
			want: []string{"aaa"},
		},
		{
			name: "Bool type",
			args: args{
				fa: &Flag[Bool]{
					Name: "f1",
				},
				value: "true",
			},
			want: true,
		},
		{
			name: "Timestamp type",
			args: args{
				fa: &Flag[TimeStamp]{
					Name: "f1",
				},
				value: "06/12/2023 03:04:05 PM",
			},
			want: tm,
		},
		{
			name: "[]Timestamp type",
			args: args{
				fa: &Flag[[]TimeStamp]{
					Name: "f1",
				},
				value: "06/12/2023 03:04:05 PM",
			},
			want: []time.Time{tm},
		},
		{
			name: "Email",
			args: args{
				fa: &Flag[Email]{
					Name: "f1",
				},
				value: "test@gmail.com",
			},
			want: "test@gmail.com",
		},
		{
			name: "File",
			args: args{
				fa: &Flag[File]{
					Name: "f1",
				},
				value: "./utils.go",
			},
			want: "./utils.go",
		},
		{
			name: "File",
			args: args{
				fa: &Flag[File]{
					Name: "f1",
				},
				value: "./*.go",
			},
			want: "./*.go",
		},
		{
			name: "IP",
			args: args{
				fa: &Flag[IP]{
					Name: "f1",
				},
				value: IP_str,
			},
			want: ip,
		},
		{
			name: "IP",
			args: args{
				fa: &Flag[[]IP]{
					Name: "f1",
				},
				value: IP_str,
			},
			want: []net.IP{ip},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.fa.SetValue(tt.args.value.(string))
			got := getFlagArgValue(tt.args.fa)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFlagArgValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
