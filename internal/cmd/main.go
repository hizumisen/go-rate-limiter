package main

import (
	"fmt"

	"github.com/hizumisen/go-rate-limiter/internal/benchmark"
)

func main() {
	fmt.Println("start")
	defer fmt.Println("finish")

	benchmark.Run()
}
