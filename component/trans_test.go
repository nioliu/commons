package component

import (
	"log"
	"reflect"
	"testing"
)

func TestTransType(t *testing.T) {
	transType, err := TransType("12342.0000", reflect.Int64)
	if err != nil {
		log.Fatal(err)
	}
	t.Log(transType)
}
