package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var nextID int = 1
var mu sync.Mutex
var orders = make(map[int]Order)


type Order struct {
	ID       int    `json:"id"`
	Product  string `json:"product"`
	Quantity int    `json:"quantity"`
}

func logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        fmt.Println(r.Method, r.URL.Path, "took", time.Since(start))
    })
}

func postHandler(w http.ResponseWriter, r *http.Request) {
    var order Order
    err := json.NewDecoder(r.Body).Decode(&order)
    if err != nil {
        http.Error(w, "errors", http.StatusBadRequest)
        return
    }
    if !isValidOrder(order.Product, order.Quantity) {
        http.Error(w, "invalid order", http.StatusBadRequest)
        return
    }

    mu.Lock()
    defer mu.Unlock()

    order.ID = nextID
    nextID++
    orders[order.ID] = order

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    errJs := json.NewEncoder(w).Encode(order)
    if errJs != nil {
        http.Error(w, "json encode errors", http.StatusBadRequest)
        return
    }
}

func getHandler(w http.ResponseWriter, r *http.Request) {
    mu.Lock()
    defer mu.Unlock()

    var resOrder []Order
    for _, ord := range orders {
        resOrder = append(resOrder, ord)
    }

    w.Header().Set("Content-Type", "application/json")
    err := json.NewEncoder(w).Encode(resOrder)
    if err != nil {
        http.Error(w, "errors json encoding", http.StatusBadRequest)
        return
    }
}

func pathHandler(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
        http.Error(w, "errors get id", http.StatusBadRequest)
        return
    }

    mu.Lock()
    defer mu.Unlock()
    order, found := orders[id]
    if !found {
        http.Error(w, "errors not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    errJson := json.NewEncoder(w).Encode(order)
    if errJson != nil {
        http.Error(w, "errors json encoding", http.StatusBadRequest)
        return
    }
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
        http.Error(w, "errors get id", http.StatusBadRequest)
        return
    }

    mu.Lock()
    defer mu.Unlock()
    _, found := orders[id]
    if !found {
        http.Error(w, "errors not found", http.StatusNotFound)
        return
    }

    delete(orders, id)

    fmt.Fprintf(w, "Order %d delete\n", id)
}

func putHandler(w http.ResponseWriter, r *http.Request) {
    var order Order
    id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
        http.Error(w, "Errors get id", http.StatusBadRequest)
        return
    }

    errDecode := json.NewDecoder(r.Body).Decode(&order)
    if errDecode != nil {
        http.Error(w, "json decode errors", http.StatusBadRequest)
        return
    }

    if !isValidOrder(order.Product, order.Quantity) {
        http.Error(w, "invalid order", http.StatusBadRequest)
        return
    }

    mu.Lock()
    defer mu.Unlock()

    order.ID = id
    _, found := orders[id]
    if !found {
        http.Error(w, "not found order with id", http.StatusNotFound)
        return
    }
    orders[id] = order

    w.Header().Set("Content-Type", "application/json")
    errJsPut := json.NewEncoder(w).Encode(order)
    if errJsPut != nil {
        http.Error(w, "json encode errors", http.StatusBadRequest)
        return
    } 
}

func isValidOrder(product string, quantity int) bool {
    if product == "" {
        return false
    }
    if quantity <= 0 {
        return false
    }
    return true
}

func main() {
    mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", postHandler)
    mux.HandleFunc("GET /orders", getHandler)
    mux.HandleFunc("GET /orders/{id}", pathHandler)
    mux.HandleFunc("DELETE /orders/{id}", deleteHandler)
    mux.HandleFunc("PUT /orders/{id}", putHandler)
	err := http.ListenAndServe(":8080", logging(mux))
	if err != nil {
		fmt.Println("Errors server", err)
		return
	}
}
