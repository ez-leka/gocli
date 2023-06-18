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

func TestHex_FromString(t *testing.T) {
	type args struct {
		v  string
		fa IFlagArg
	}
	tests := []struct {
		name    string
		s       *Hex
		args    args
		wantErr bool
	}{
		{
			name: "t1",
			s:    new(Hex),
			args: args{
				v: "7B",
				fa: &Flag[Hex]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t2",
			s:    new(Hex),
			args: args{
				v: "007B",
				fa: &Flag[Hex]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t3",
			s:    new(Hex),
			args: args{
				v: "7B",
				fa: &Flag[Hex]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t4",
			s:    new(Hex),
			args: args{
				v: "7B",
				fa: &Flag[Hex]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t5",
			s:    new(Hex),
			args: args{
				v: "0x7B",
				fa: &Flag[Hex]{
					Name: "test",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.FromString(tt.args.v, tt.args.fa); (err != nil) != tt.wantErr {
				t.Errorf("Hex.FromString() error = %v, wantErr %v", err, tt.wantErr)
			}
			i := tt.s.GetValue().(int)
			if i != 123 {
				t.Errorf("Hex.FromString() wrong value: wanted 123, got %d", i)
			}
		})
	}
}

func TestOctal_FromString(t *testing.T) {
	type args struct {
		v  string
		fa IFlagArg
	}
	tests := []struct {
		name    string
		s       *Octal
		args    args
		wantErr bool
	}{
		{
			name: "t1",
			s:    new(Octal),
			args: args{
				v: "173",
				fa: &Flag[Hex]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t2",
			s:    new(Octal),
			args: args{
				v: "183",
				fa: &Flag[Hex]{
					Name: "test",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if err = tt.s.FromString(tt.args.v, tt.args.fa); (err != nil) != tt.wantErr {
				t.Errorf("Octal.FromString() error = %v, wantErr %v", err, tt.wantErr)
			}
			i := tt.s.GetValue().(int)
			if err == nil && i != 123 {
				t.Errorf("Octal.FromString() wrong value: wanted 123, got %d", i)
			}
		})
	}
}

func TestBinary_FromString(t *testing.T) {
	type args struct {
		v  string
		fa IFlagArg
	}
	tests := []struct {
		name    string
		s       *Binary
		args    args
		wantErr bool
	}{
		{
			name: "t1",
			s:    new(Binary),
			args: args{
				v: "1111011",
				fa: &Flag[Hex]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t2",
			s:    new(Binary),
			args: args{
				v: "0000000001111011",
				fa: &Flag[Hex]{
					Name: "test",
				},
			},
			wantErr: false,
		},
		{
			name: "t3",
			s:    new(Binary),
			args: args{
				v: "123",
				fa: &Flag[Hex]{
					Name: "test",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			if err = tt.s.FromString(tt.args.v, tt.args.fa); (err != nil) != tt.wantErr {
				t.Errorf("Binary.FromString() error = %v, wantErr %v", err, tt.wantErr)
			}
			i := tt.s.GetValue().(int)
			if err == nil && i != 123 {
				t.Errorf("Binary.FromString() wrong value: wanted 123, got %d", i)
			}
		})
	}
}
