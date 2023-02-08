package errs

import "testing"

func TestErr(t *testing.T) {
	var LackRequiredParametersError = NewError(10000, "lack required parameters")
	s := LackRequiredParametersError.Error()
	t.Log(s)
	newError := NewError(0, "qwe")
	println(newError.WithDescription("321").Error())
}
