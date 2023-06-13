package gocli

import (
	"testing"
)

func TestTimeStamp_FromString(t *testing.T) {
	type args struct {
		v  string
		fa IFlagArg
	}
	tests := []struct {
		name    string
		s       *TimeStamp
		args    args
		wantErr bool
	}{
		{
			name: "t1",
			s:    new(TimeStamp),
			args: args{
				v: "02 Feb 06 15:04 MST",
				fa: &Flag[TimeStamp]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t2",
			s:    new(TimeStamp),
			args: args{
				v: "02 Jan 06 15:04 -0700",
				fa: &Flag[TimeStamp]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t3",
			s:    new(TimeStamp),
			args: args{
				v: "Monday, 02-Feb-06 15:04:05 MST",
				fa: &Flag[TimeStamp]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t4",
			s:    new(TimeStamp),
			args: args{
				v: "Mon, 02 Feb 2023 15:04:05 MST",
				fa: &Flag[TimeStamp]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t5",
			s:    new(TimeStamp),
			args: args{
				v: "Mon, 02 Jan 2006 15:04:05 -0700",
				fa: &Flag[TimeStamp]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t6",
			s:    new(TimeStamp),
			args: args{
				v: "2006-01-02T15:04:05Z",
				fa: &Flag[TimeStamp]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t7",
			s:    new(TimeStamp),
			args: args{
				v: "2006-01-02T15:04:05.999999999Z",
				fa: &Flag[TimeStamp]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t8",
			s:    new(TimeStamp),
			args: args{
				v: "01/02/2020 01:02:00 PM",
				fa: &Flag[TimeStamp]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t9",
			s:    new(TimeStamp),
			args: args{
				v: "04:00PM",
				fa: &Flag[TimeStamp]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t10",
			s:    new(TimeStamp),
			args: args{
				v: "04:00 AM",
				fa: &Flag[TimeStamp]{
					Name: "test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			s := &TimeStamp{}
			if err := s.FromString(tt.args.v, tt.args.fa); (err != nil) != tt.wantErr {
				t.Errorf("TimeStamp.FromString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
