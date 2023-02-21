package component

import "fmt"

type app struct {
	name string
}

func (a *app) Hello() {
	fmt.Println("hello!!!")
	fmt.Println(a.name)

}
