package str

import (
	"reflect"
	"testing"
)

func TestSplitAny(t *testing.T) {
	type args struct {
		s   string
		sep []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "",
			args: args{
				s:   "((cpe:2.3:o:suse:open_suse:10.3:*:*:*:*:*:*:*) $AND$ cpe:2.3:o:suse:open_suse:10.3:*:*:*:*:*:*:* $OR$ cpe:2.3:o:suse:open_suse:11.0:*:*:*:*:*:*:*)",
				sep: []string{" ", "(", ")", "$OR$", "$AND$"}},
			want: []string{
				"cpe:2.3:o:suse:open_suse:10.3:*:*:*:*:*:*:*",
				"cpe:2.3:o:suse:open_suse:10.3:*:*:*:*:*:*:*",
				"cpe:2.3:o:suse:open_suse:11.0:*:*:*:*:*:*:*",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitAny(tt.args.s, tt.args.sep...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitAny() = %v, want %v", got, tt.want)
			}
		})
	}
}
