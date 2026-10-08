package main

import (
	"testing"
	"time"
)

func TestLimitadorBloqueiaDepoisDaRajadaEReabastece(t *testing.T) {
	l := &limitador{baldes: make(map[string]*balde)}
	agora := time.Now()

	for i := 0; i < int(rajada); i++ {
		if !l.permitir("1.1.1.1", agora) {
			t.Fatalf("requisição %d deveria passar (dentro da rajada)", i+1)
		}
	}
	if l.permitir("1.1.1.1", agora) {
		t.Fatal("depois da rajada, a requisição deveria ser bloqueada")
	}
	if !l.permitir("2.2.2.2", agora) {
		t.Fatal("outro visitante não deve ser afetado")
	}
	if !l.permitir("1.1.1.1", agora.Add(time.Second)) {
		t.Fatal("após 1s o balde deveria ter reabastecido")
	}
}
