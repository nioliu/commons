package component

import (
	"log"
	"reflect"
	"testing"
)

func TestTransType(t *testing.T) {
	transType, err := TransType(int8(3), reflect.Int32)
	if err != nil {
		log.Fatal(err)
	}
	t.Log(transType)
}
