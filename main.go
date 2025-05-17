package main

import (
	"fmt"
	"net/http"
	"sync"
)

func Hello(str string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello %v", str)
	}
}

// Тест для Balancer
func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		mux1 := http.NewServeMux()
		mux1.HandleFunc("/", Hello("test1"))

		fmt.Println("Started server on 8081")
		if err := http.ListenAndServe(":8081", mux1); err != nil {
			fmt.Printf("Server 1 error: %v\n", err)
		}
	}()

	go func() {
		defer wg.Done()
		mux2 := http.NewServeMux()
		mux2.HandleFunc("/", Hello("test2"))

		fmt.Println("Started server on 8082")
		if err := http.ListenAndServe(":8082", mux2); err != nil {
			fmt.Printf("Server 2 error: %v\n", err)
		}
	}()

	wg.Wait()
}
