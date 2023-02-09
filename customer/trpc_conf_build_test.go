package customer

import (
	"log"
	"os"
	"testing"
)

func TestTransTrpcYamlConfToLocal(t *testing.T) {
	file, err := os.Open("trpc_go.yaml")
	if err != nil {
		log.Fatalln(err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalln(err)
	}
	bytes := make([]byte, fileInfo.Size())
	n, err := file.Read(bytes)
	if err != nil {
		log.Fatalln(err)
	}
	t.Log("read: ", n)
	t.Log(string(bytes))

	//local, err := TransTrpcYamlConfToLocal(bytes)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//t.Log(local)
}
