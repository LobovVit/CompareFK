package compare

import (
	"math/rand"
	"testing"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func Test_intersection(t *testing.T) {
	type args struct {
		master []string
		slave  []string
	}
	m := make([]string, 100000)
	s := make([]string, 100000)
	for i := 1; i < 100000; i++ {
		m = append(m, RandStringRunes(3))
		s = append(s, RandStringRunes(3))
	}
	tests := []struct {
		name     string
		args     args
		wantDiff []string
	}{
		{name: "tesr1",
			args:     args{master: m, slave: s},
			wantDiff: m},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDiff := intersection(tt.args.master, tt.args.slave)
			t.Log(gotDiff)
		})
	}
}
