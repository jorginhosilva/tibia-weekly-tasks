package main

import (
	"strings"
	"testing"
)

func TestInterpretarArgumentos(t *testing.T) {
	casos := []struct {
		nome       string
		args       []string
		modo       modoExecucao
		confirmado bool
		erro       string // trecho esperado no erro (vazio = não deve haver erro)
	}{
		{"sem argumentos: demonstração", nil, modoDemo, false, ""},
		{"web", []string{"web"}, modoWeb, false, ""},
		{"recriar sem confirmação", []string{"recriar"}, modoRecriar, false, ""},
		{"recriar com confirmação", []string{"recriar", "--confirmo"}, modoRecriar, true, ""},
		{"ajuda", []string{"ajuda"}, modoAjuda, false, ""},
		{"-h", []string{"-h"}, modoAjuda, false, ""},
		{"--help", []string{"--help"}, modoAjuda, false, ""},
		// O erro de digitação não pode cair no modo que grava no banco:
		{"web escrito errado", []string{"wbe"}, "", false, "modo desconhecido"},
		{"modo inexistente", []string{"apagar"}, "", false, "modo desconhecido"},
		{"web com argumento extra", []string{"web", "x"}, "", false, "não aceita argumentos"},
		{"recriar com flag errada", []string{"recriar", "--sim"}, "", false, "argumento desconhecido"},
		{"recriar com argumentos demais", []string{"recriar", "--confirmo", "x"}, "", false, "argumento desconhecido"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			modo, confirmado, err := interpretarArgumentos(c.args)
			if c.erro != "" {
				if err == nil || !strings.Contains(err.Error(), c.erro) {
					t.Fatalf("erro = %v, esperava um erro com %q", err, c.erro)
				}
				if !strings.Contains(err.Error(), "Uso:") {
					t.Errorf("a mensagem de erro deveria mostrar como usar o programa")
				}
				return
			}
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if modo != c.modo || confirmado != c.confirmado {
				t.Errorf("obtido (%q, %v), esperava (%q, %v)", modo, confirmado, c.modo, c.confirmado)
			}
		})
	}
}

func TestCarregarConfigBanco(t *testing.T) {
	// Numa pasta vazia, para o .env REAL do projeto não ser lido durante o teste.
	t.Chdir(t.TempDir())

	definir := func(usuario, host, porta, banco string) {
		t.Setenv("DB_USER", usuario)
		t.Setenv("DB_HOST", host)
		t.Setenv("DB_PORT", porta)
		t.Setenv("DB_NAME", banco)
		t.Setenv("DB_PASSWORD", "s3nh@/com:simbolos")
		t.Setenv("DB_TLS", "")
	}

	t.Run("cita todas as variáveis que faltam, em ordem", func(t *testing.T) {
		definir("", "127.0.0.1", "", "")
		_, err := carregarConfigBanco()
		if err == nil || !strings.Contains(err.Error(), "DB_USER, DB_PORT, DB_NAME") {
			t.Errorf("erro = %v, esperava citar DB_USER, DB_PORT, DB_NAME nessa ordem", err)
		}
	})

	t.Run("configuração completa", func(t *testing.T) {
		definir("leitura", "db.exemplo.com", "3306", "tibia")
		cfg, err := carregarConfigBanco()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.User != "leitura" || cfg.Addr != "db.exemplo.com:3306" || cfg.DBName != "tibia" {
			t.Errorf("configuração inesperada: %+v", cfg)
		}
		if cfg.Passwd != "s3nh@/com:simbolos" {
			t.Error("a senha (com símbolos) deveria passar intacta")
		}
		if cfg.TLS != nil {
			t.Error("sem DB_TLS=true, a conexão não deveria exigir TLS")
		}
	})

	t.Run("DB_TLS=true liga a criptografia", func(t *testing.T) {
		definir("leitura", "db.exemplo.com", "3306", "tibia")
		t.Setenv("DB_TLS", "true")
		cfg, err := carregarConfigBanco()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.TLS == nil || cfg.TLS.ServerName != "db.exemplo.com" {
			t.Errorf("TLS = %+v, esperava ServerName db.exemplo.com", cfg.TLS)
		}
	})
}
