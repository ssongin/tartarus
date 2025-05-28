package main

import (
	"log"
	"net/http"

	"github.com/ssongin/tartarus/api"
)

func main() {
	log.Println("Starting server at http://localhost:8080")
	http.HandleFunc("/compress", api.HandleCompress)
	http.HandleFunc("/decompress", api.HandleDecompress)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
