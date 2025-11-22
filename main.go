package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println(" 🔨 Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
