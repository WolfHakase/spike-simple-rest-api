package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_getItemHandler(t *testing.T) {
	path := fmt.Sprintf("/items/%d", 1)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/items/{id}", getItem)
	router.ServeHTTP(rr, req)

	// integration test
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// contract test
	expected := `{"id":1,"name":"second","description":"second item"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func Test_listItemsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/items/?filter=second", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/items/", listItems).Queries("filter", "{filter}")
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `[{"id":1,"name":"second","description":"second item"}]`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func Test_createItemHandler(t *testing.T) {
	newItem := []byte(`{"name":"new_name","description":"new_description"}`)

	req, err := http.NewRequest("POST", "/items/", bytes.NewBuffer(newItem))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/items/", createItem)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	expected := `{"id":2,"name":"new_name","description":"new_description"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
