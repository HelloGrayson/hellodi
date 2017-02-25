package main

import (
	"fmt"

	"github.com/breerly/hellodi/fx2"
)

func main() {
	service := fx2.New()

	fmt.Println("hellodi", service)
}
