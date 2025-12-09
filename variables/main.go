package main

import (
	"fmt"
)

func main() {
	var myIntVar int
	myIntVar = -12
	fmt.Println("El valor de myIntVar es:", myIntVar)

	var myUintVar uint
	myUintVar = 12
	fmt.Println("El valor de myUintVar es:", myUintVar)

	var myFloatVar float32
	myFloatVar = 3.14
	fmt.Println("El valor de myFloatVar es:", myFloatVar)

	var myStringVar string
	myStringVar = "Hola, Go!"
	fmt.Println("El valor de myStringVar es:", myStringVar)
	fmt.Println("La dirección de myStringVar es:", &myStringVar)

	var myBoolVar bool
	myBoolVar = true
	fmt.Println("El valor de myBoolVar es:", myBoolVar)

	myIntVar2 := 12
	fmt.Println("El valor de myIntVar2 es:", myIntVar2)

	const myFirstConst = "Constante en Go"
	fmt.Println("El valor de myFirstConst es:", myFirstConst)

	const myIntConst int = 100
	fmt.Println("El valor de myIntConst es:", myIntConst)

}
