package component

import "testing"

func Test_app_Hello(t *testing.T) {
	a := (*app)(nil)
	a.Hello()
}
