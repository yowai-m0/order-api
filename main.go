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
        if order.Product == "" {
            http.Error(w, "product is required", http.StatusBadRequest)
            return
        }
        if order.Quantity <= 0 {
            http.Error(w, "quantity must be positive", http.StatusBadRequest)
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

    case http.MethodGet:
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

    if  order.Product == "" {
        http.Error(w, "product is required", http.StatusBadRequest)
        return
    }
    if order.Quantity <= 0 {
        http.Error(w, "quantity must be positive", http.StatusBadRequest)
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
	http.HandleFunc("/orders", createHandlerOrders)
    http.HandleFunc("GET /orders/{id}", pathHandler)
    http.HandleFunc("DELETE /orders/{id}", deleteHandler)
    http.HandleFunc("PUT /orders/{id}", putHandler) 
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Errors server", err)
		return
	}
}
