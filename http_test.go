package stunning

import (
	"testing"
	"fmt"
	"net/http"
	"log"
)

func TestHttp(t *testing.T){
	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		log.Printf("%s", r.RemoteAddr)
		fmt.Fprintf(w, "Welcome to my website!")
	})

	log.Println(http.ListenAndServe(":8080", nil))
}
