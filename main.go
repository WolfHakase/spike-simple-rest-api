package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var (
	NotFoundError = errors.New("not found")
)

type PingResponse struct {
	Ping string
}

type Item struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var itemRepository = []Item{
	{
		ID:          0,
		Name:        "first",
		Description: "first item",
	},
	{
		ID:          1,
		Name:        "second",
		Description: "second item",
	},
}

func main() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/ping", ping).Methods(http.MethodGet)
	itemRoutes := r.PathPrefix("/items").Subrouter()
	itemRoutes.HandleFunc("/{id}/duplicate", duplicateItem).Methods(http.MethodPost, http.MethodOptions)
	itemRoutes.HandleFunc("/{id}", getItem).Methods(http.MethodGet, http.MethodOptions)
	itemRoutes.HandleFunc("/{id}", deleteItem).Methods(http.MethodDelete, http.MethodOptions)
	itemRoutes.HandleFunc("/{id}", updateItem).Methods(http.MethodPut, http.MethodOptions)
	itemRoutes.HandleFunc("/", createItem).Methods(http.MethodPost, http.MethodOptions)
	itemRoutes.HandleFunc("/", listItems).Methods(http.MethodGet, http.MethodOptions)
	itemRoutes.HandleFunc("/", routeDoesNotExist)
	r.Use(loggingMiddleware)
	r.Use(mux.CORSMethodMiddleware(r))

	srv := &http.Server{
		Addr:         "0.0.0.0:8000",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go log.Fatal(srv.ListenAndServe())

	waitUntilShutdown()
	gracefulShutdown(srv, wait)
}

func waitUntilShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func gracefulShutdown(srv *http.Server, wait time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}

func ping(w http.ResponseWriter, r *http.Request) {
	SuccessResponse(w, PingResponse{Ping: "Pong"})
}

func listItems(w http.ResponseWriter, r *http.Request) {
	SuccessResponse(w, itemRepository)
}

func getItem(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r)
	if err != nil {
		BadRequestResponse(w, "invalid ID")
		return
	}

	item, err := getItemForID(*id)
	if err != nil {
		NotFoundResponse(w, "item with ID does not exist")
		return
	}

	SuccessResponse(w, item)
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r)
	if err != nil {
		BadRequestResponse(w, "invalid ID")
		return
	}

	err = doDeleteItem(*id)
	if err != nil {
		NotFoundResponse(w, "item with ID does not exist")
		return
	}

	NoContentResponse(w)
}

func duplicateItem(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r)
	if err != nil {
		BadRequestResponse(w, "invalid ID")
		return
	}

	item, err := getItemForID(*id)
	if err != nil {
		NotFoundResponse(w, "item with ID does not exist")
		return
	}

	duplicate := Item{
		ID:          getNextID(),
		Name:        item.Name,
		Description: item.Description,
	}
	err = doCreateItem(duplicate)
	if err != nil {
		InternalErrorResponse(w, "could not duplicate item")
		return
	}

	CreatedResponse(w, duplicate)
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	id, err := getIDParam(r)
	if err != nil {
		BadRequestResponse(w, "invalid ID")
		return
	}

	var item Item
	err = decodeBody(r, &item)
	if err != nil {
		BadRequestResponse(w, "could not decode request body")
	}

	item.ID = *id

	err = doUpdateItem(item)
	if err != nil {
		InternalErrorResponse(w, "could not update item")
	}

	SuccessResponse(w, item)
}

func createItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	err := decodeBody(r, &item)
	if err != nil {
		BadRequestResponse(w, "could not decode request body")
	}

	item.ID = getNextID()

	err = doCreateItem(item)
	if err != nil {
		InternalErrorResponse(w, "could not create item")
	}

	CreatedResponse(w, item)
}

func routeDoesNotExist(w http.ResponseWriter, r *http.Request) {
	NotFoundResponse(w, "endpoint does not exist")
}

func decodeBody(r *http.Request, target interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(target)
	if err != nil {
		return err
	}
	err = r.Body.Close()
	if err != nil {
		return err
	}
	return nil
}

func getIDParam(r *http.Request) (*int, error) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func getItemForID(id int) (*Item, error) {
	for _, item := range itemRepository {
		if item.ID == id {
			return &item, nil
		}
	}
	return nil, NotFoundError
}

func doCreateItem(item Item) error {
	itemRepository = append(itemRepository, item)
	return nil
}

func doDeleteItem(id int) error {
	var index *int
	for i, item := range itemRepository {
		if item.ID == id {
			index = &i
			break
		}
	}

	if index == nil {
		return NotFoundError
	}
	itemRepository = append(itemRepository[:*index], itemRepository[*index+1:]...)
	return nil
}

func doUpdateItem(updateValues Item) error {
	var index *int
	for i, item := range itemRepository {
		if item.ID == updateValues.ID {
			index = &i
			break
		}
	}

	if index == nil {
		return NotFoundError
	}
	itemRepository = append(itemRepository[:*index], updateValues)
	itemRepository = append(itemRepository, itemRepository[*index+1:]...)

	return nil
}

func getNextID() int {
	return len(itemRepository)
}

func SuccessResponse(w http.ResponseWriter, payload interface{}) {
	JSONResponse(w, http.StatusOK, payload)
}

func CreatedResponse(w http.ResponseWriter, payload interface{}) {
	JSONResponse(w, http.StatusCreated, payload)
}

func InternalErrorResponse(w http.ResponseWriter, message string) {
	JSONResponse(w, http.StatusInternalServerError, map[string]string{"error": message})
}

func BadRequestResponse(w http.ResponseWriter, message string) {
	JSONResponse(w, http.StatusBadRequest, map[string]string{"error": message})
}

func NotFoundResponse(w http.ResponseWriter, message string) {
	JSONResponse(w, http.StatusNotFound, map[string]string{"error": message})
}

func NoContentResponse(w http.ResponseWriter) {
	JSONResponse(w, http.StatusNoContent, map[string]string{})
}

func JSONResponse(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
