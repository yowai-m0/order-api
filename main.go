package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

var nextID int = 1
var mu sync.Mutex
var orders = make(map[int]Order)


type Order struct {
	ID       int    `json:"id"`
	Product  string `json:"product"`
	Quantity int    `json:"quantity"`
}

func createHandlerOrders(w http.ResponseWriter, r *http.Request) {
    var order Order
    switch r.Method {
    case http.MethodPost:
        err := json.NewDecoder(r.Body).Decode(&order)
        if err != nil {
            http.Error(w, "errors", http.StatusBadRequest)
            return
        }

        mu.Lock()
        defer mu.Unlock()

        order.ID = nextID
        nextID++
        orders[order.ID] = order

        fmt.Fprintf(w, "Create a new order\n  ID: %d\n  Product: %s (quantity: %d)\n", 
        order.ID, order.Product, order.Quantity)

    case http.MethodGet:
        mu.Lock()
        defer mu.Unlock()

        var resOrder []Order
        for _, ord := range orders {
            resOrder = append(resOrder, ord)
        }

        w.Header().Set("Content-Type", "application/json")
        err := json.NewEncoder(w).Encode(&resOrder)
        if err != nil {
            http.Error(w, "errors json encoding", http.StatusBadRequest)
            return
        }

    default:
        http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
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

func main() {
	http.HandleFunc("/orders", createHandlerOrders)
    http.HandleFunc("GET /orders/{id}", pathHandler)
    http.HandleFunc("DELETE /orders/{id}", deleteHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Errors server", err)
		return
	}
}
