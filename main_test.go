package main

import "testing"

func TestValidOrder(t *testing.T) {
    if isValidOrder("", 5) {
		t.Errorf("ожидалось false для пустого product, получили true")
	}

	if isValidOrder("Ноутбук", 0) {
		t.Errorf("ожидалось false для quantity = 0, получили true")
	}

	if isValidOrder("Ноутбук", -3) {
		t.Errorf("ожидалось false для отрицательного quantity, получили true")
	}

	if !isValidOrder("Ноутбук", 2) {
		t.Errorf("ожидалось true для корректных данных, получили false")
	}
}