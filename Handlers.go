package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.Dir("./wwwroot")).ServeHTTP(w, r)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("c")
	if key == "" {
		key = "a"
	}
	AllSentences, err := SentencesLoad(key)
	if err != nil {
		log.Fatal(err)
		return
	}
	num := new(rand.Intn(len(AllSentences)))
	S := AllSentences[*num]
	//print(num)
	payload, err := json.Marshal(S)
	if err != nil {
		log.Fatal(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(payload)
	if err != nil {
		log.Fatal(err)
		return
	}
	return
}
