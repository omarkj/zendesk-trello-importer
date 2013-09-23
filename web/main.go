package main

import (
  "os"
  "fmt"
  "net/http"
)

var port string = os.Getenv("PORT")

func main() {
    fmt.Println("starting webserver on port", port)
    http.HandleFunc("/", handler)
    http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

