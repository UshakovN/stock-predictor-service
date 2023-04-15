package utils

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func HandleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("/ok"))
		if err != nil {
			log.Printf("cannot write to response writer: %v", err)
		}
	}
}

func ContinuouslyServe(port string) {
	err := http.ListenAndServe(fmt.Sprint(":", port), nil)
	if err != nil {
		log.Fatalf("listen and serve error: %v", err)
	}
}
