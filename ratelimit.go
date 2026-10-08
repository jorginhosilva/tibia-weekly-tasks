package main

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Limite de requisições por visitante (IP). Protege o banco e o servidor de
// robôs e de quem tentar derrubar o site com milhares de buscas por segundo.
//
// Funciona como um "balde de fichas": cada visitante começa com `rajada`
// fichas, cada requisição gasta uma, e o balde reabastece `porSegundo` fichas
// por segundo. Sem ficha, a resposta é 429 (Too Many Requests).
const (
	porSegundo = 10.0 // ritmo contínuo permitido (o autocomplete usa ~5/s ao digitar)
	rajada     = 30.0 // pico permitido de uma vez
	ociosoMax  = 5 * time.Minute
)

type balde struct {
	fichas    float64
	ultimaVez time.Time
}

type limitador struct {
	mu     sync.Mutex
	baldes map[string]*balde
}

func novoLimitador() *limitador {
	l := &limitador{baldes: make(map[string]*balde)}
	go l.limparPeriodicamente()
	return l
}

// permitir diz se o visitante `ip` ainda pode fazer uma requisição agora.
func (l *limitador) permitir(ip string, agora time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, existe := l.baldes[ip]
	if !existe {
		l.baldes[ip] = &balde{fichas: rajada - 1, ultimaVez: agora}
		return true
	}

	// Reabastece conforme o tempo que passou, sem passar do máximo.
	b.fichas += agora.Sub(b.ultimaVez).Seconds() * porSegundo
	if b.fichas > rajada {
		b.fichas = rajada
	}
	b.ultimaVez = agora

	if b.fichas < 1 {
		return false
	}
	b.fichas--
	return true
}

// limparPeriodicamente apaga visitantes inativos, para o mapa não crescer para sempre.
func (l *limitador) limparPeriodicamente() {
	for range time.Tick(time.Minute) {
		limite := time.Now().Add(-ociosoMax)
		l.mu.Lock()
		for ip, b := range l.baldes {
			if b.ultimaVez.Before(limite) {
				delete(l.baldes, ip)
			}
		}
		l.mu.Unlock()
	}
}

// limitarRequisicoes aplica o limite apenas às rotas /api/.
func limitarRequisicoes(next http.Handler) http.Handler {
	l := novoLimitador()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") && !l.permitir(ipDoCliente(r), time.Now()) {
			w.Header().Set("Retry-After", "1")
			escreverErro(w, http.StatusTooManyRequests, "Muitas requisições. Aguarde um instante e tente de novo.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ipDoCliente descobre o IP de quem fez a requisição.
//
// Atrás de uma hospedagem (Railway, Render, Fly...), r.RemoteAddr é o IP do
// proxy da plataforma, igual para todo mundo, e o limite pegaria todos juntos.
// Nesse caso defina TRUST_PROXY=true: usamos o ÚLTIMO IP de X-Forwarded-For,
// que é o que o proxy da plataforma acrescentou (os anteriores o visitante
// pode forjar). Só ative se o site realmente estiver atrás de um proxy.
func ipDoCliente(r *http.Request) string {
	if os.Getenv("TRUST_PROXY") == "true" {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			partes := strings.Split(xff, ",")
			if ip := strings.TrimSpace(partes[len(partes)-1]); ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
