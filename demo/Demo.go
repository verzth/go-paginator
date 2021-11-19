package main

import (
	"fmt"
	"git.teknoku.digital/teknoku/jumper"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", Index).Methods("GET")

	err := http.ListenAndServe(":9999", handlers.CORS(
		handlers.AllowedHeaders([]string{"Content-Type","Authorization"}),
		handlers.AllowedMethods([]string{http.MethodGet}),
		handlers.AllowedOrigins([]string{"*"}),
	)(r))
	if err != nil {
		panic(err)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	//var req = jumper.PlugRequest(r, w)
	var res = jumper.PlugResponse(w)

	res.ReplySuccess("0000000", "SSSSSS", "Success", fmt.Sprintf("%s%s?page=%d", r.Host, r.URL.Path, 1))
}