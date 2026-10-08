package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	tamanhoMaxLog   = 64 << 10 // 64 KB: bem mais do que qualquer log real
	tempoMaxRequest = 10 * time.Second

	// O JSON "engorda" o texto (cada quebra de linha vira \n, aspas viram \"),
	// então o corpo da requisição pode ser maior que o log em si.
	tamanhoMaxCorpoLog = 8 * tamanhoMaxLog

	// validadeSaude é por quanto tempo o resultado do /healthz é reaproveitado,
	// para que chamadas em sequência não sobrecarreguem o banco.
	validadeSaude = 5 * time.Second

	// pastaSite é onde o `npm run build` do front-end coloca o site pronto.
	// Pode ser trocada com a variável de ambiente SITE_DIR.
	pastaSite = "frontend/dist"
)

type servidor struct {
	db   *sql.DB
	site string // pasta com o site compilado

	// Resultado guardado do último teste de saúde do banco (/healthz).
	saudeMu  sync.Mutex
	saudeEm  time.Time
	saudeErr error
}

// NovoServidorHTTP monta as rotas (endereços): a API em /api/... e o site
// (arquivos do front-end) em todo o resto.
func NovoServidorHTTP(db *sql.DB, addr string) *http.Server {
	pasta := pastaSite
	if p := os.Getenv("SITE_DIR"); p != "" {
		pasta = p
	}
	s := &servidor{db: db, site: pasta}

	return &http.Server{
		Addr:              addr,
		Handler:           s.rotas(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// rotas liga cada endereço à função que o atende, já com os cabeçalhos de
// segurança e o limite de requisições. Fica separada de NovoServidorHTTP
// para os testes poderem usá-la sem abrir uma porta de verdade.
func (s *servidor) rotas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/buscar", s.apiBuscar)
	mux.HandleFunc("GET /api/sugestoes", s.apiSugestoes)
	mux.HandleFunc("POST /api/log", s.apiLog)
	mux.HandleFunc("/api/", apiNaoEncontrada) // qualquer outro /api/...: resposta em JSON, nunca a página do site
	mux.HandleFunc("GET /healthz", s.saude)
	mux.HandleFunc("/", s.servirSite) // o site: só GET e HEAD (conferido dentro da função)
	return cabecalhosSeguranca(limitarRequisicoes(mux))
}

// cabecalhosSeguranca adiciona avisos ao navegador para ele ser mais rígido:
// só executa scripts e estilos que vêm do nosso próprio site.
func cabecalhosSeguranca(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy",
			"default-src 'self'; img-src 'self' data:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Desliga recursos do navegador que o site não usa (câmera, GPS...).
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		// HSTS: depois da primeira visita, o navegador só acessa o site por HTTPS.
		// (É ignorado em http://localhost, então não atrapalha o desenvolvimento.)
		// Sem includeSubDomains de propósito: assim um subdomínio futuro sem HTTPS
		// não fica inacessível.
		h.Set("Strict-Transport-Security", "max-age=31536000")
		next.ServeHTTP(w, r)
	})
}

// metodoDaApi diz qual método cada endereço da API aceita. Serve para
// responder 405 (método errado) em vez de 404 quando o endereço existe.
var metodoDaApi = map[string]string{
	"/api/buscar":    http.MethodGet,
	"/api/sugestoes": http.MethodGet,
	"/api/log":       http.MethodPost,
}

// apiNaoEncontrada atende o que sobrou de /api/: método errado num endereço
// que existe (405) ou um endereço que não existe (404), sempre em JSON.
func apiNaoEncontrada(w http.ResponseWriter, r *http.Request) {
	if metodo, existe := metodoDaApi[r.URL.Path]; existe {
		w.Header().Set("Allow", metodo)
		escreverErro(w, http.StatusMethodNotAllowed, "método não permitido: use "+metodo)
		return
	}
	escreverErro(w, http.StatusNotFound, "endereço da API não encontrado")
}

// bancoSaudavel testa o banco, mas reaproveita o resultado por validadeSaude.
func (s *servidor) bancoSaudavel() error {
	s.saudeMu.Lock()
	defer s.saudeMu.Unlock()

	if !s.saudeEm.IsZero() && time.Since(s.saudeEm) < validadeSaude {
		return s.saudeErr
	}

	// Contexto próprio (e não o da requisição): se o visitante desistir no meio,
	// a falha dele não pode ficar guardada como se fosse a do banco.
	ctx, cancelar := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelar()

	s.saudeErr = s.db.PingContext(ctx)
	s.saudeEm = time.Now()
	return s.saudeErr
}

// saude é o "health check": a hospedagem chama /healthz para saber se o site
// e o banco estão de pé, e reinicia o serviço se não estiverem.
func (s *servidor) saude(w http.ResponseWriter, r *http.Request) {
	if err := s.bancoSaudavel(); err != nil {
		log.Printf("health check: banco indisponível: %v", err)
		escreverErro(w, http.StatusServiceUnavailable, "banco indisponível")
		return
	}
	escreverJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- API -------------------------------------------------------------------

// resultadoBusca é o formato que o front-end espera: um nome e a lista
// de nomes relacionados (itens, se buscou criatura; criaturas, se buscou item).
type resultadoBusca struct {
	Nome         string   `json:"nome"`
	Relacionados []string `json:"relacionados"`
}

type itemLog struct {
	Nome       string `json:"nome"`
	Quantidade int    `json:"quantidade"`
}

// GET /api/buscar?tipo=criatura|item&q=texto
func (s *servidor) apiBuscar(w http.ResponseWriter, r *http.Request) {
	ctx, cancelar := context.WithTimeout(r.Context(), tempoMaxRequest)
	defer cancelar()

	termo := r.URL.Query().Get("q")
	resultado := []resultadoBusca{} // [] (e não null) quando não houver nada
	var err error

	switch r.URL.Query().Get("tipo") {
	case "criatura":
		var achados []CriaturaComItens
		if achados, err = BuscarPorCriatura(ctx, s.db, termo); err == nil {
			for _, c := range achados {
				resultado = append(resultado, resultadoBusca{c.Criatura, c.Itens})
			}
		}
	case "item":
		var achados []ItemComCriaturas
		if achados, err = BuscarPorItem(ctx, s.db, termo); err == nil {
			for _, i := range achados {
				resultado = append(resultado, resultadoBusca{i.Item, i.Criaturas})
			}
		}
	default:
		escreverErro(w, http.StatusBadRequest, "tipo de busca inválido")
		return
	}

	if s.tratarErroBusca(w, err) {
		return
	}
	escreverJSON(w, http.StatusOK, resultado)
}

// GET /api/sugestoes?tipo=criatura|item&q=texto  (autocomplete)
func (s *servidor) apiSugestoes(w http.ResponseWriter, r *http.Request) {
	ctx, cancelar := context.WithTimeout(r.Context(), tempoMaxRequest)
	defer cancelar()

	termo := r.URL.Query().Get("q")
	var nomes []string
	var err error

	switch r.URL.Query().Get("tipo") {
	case "criatura":
		nomes, err = SugerirCriaturas(ctx, s.db, termo)
	case "item":
		nomes, err = SugerirItens(ctx, s.db, termo)
	default:
		escreverErro(w, http.StatusBadRequest, "tipo de busca inválido")
		return
	}

	// Termo vazio no autocomplete não é erro: simplesmente não há sugestões.
	if errors.Is(err, ErrBuscaVazia) {
		escreverJSON(w, http.StatusOK, []string{})
		return
	}
	if s.tratarErroBusca(w, err) {
		return
	}
	escreverJSON(w, http.StatusOK, nomes)
}

// POST /api/log  com corpo {"log": "texto do Hunt Analyser"}
func (s *servidor) apiLog(w http.ResponseWriter, r *http.Request) {
	// MaxBytesReader corta o envio se passar do limite.
	r.Body = http.MaxBytesReader(w, r.Body, tamanhoMaxCorpoLog)

	var corpo struct {
		Log string `json:"log"`
	}
	if err := json.NewDecoder(r.Body).Decode(&corpo); err != nil {
		var grande *http.MaxBytesError
		if errors.As(err, &grande) {
			escreverErro(w, http.StatusRequestEntityTooLarge, "Texto grande demais (máximo 64 KB).")
			return
		}
		// Qualquer outro erro de leitura: o JSON está malformado ou vazio.
		escreverErro(w, http.StatusBadRequest, `Requisição inválida: envie um JSON no formato {"log": "texto"}.`)
		return
	}
	if len(corpo.Log) > tamanhoMaxLog {
		escreverErro(w, http.StatusRequestEntityTooLarge, "Texto grande demais (máximo 64 KB).")
		return
	}

	ctx, cancelar := context.WithTimeout(r.Context(), tempoMaxRequest)
	defer cancelar()

	itens, err := ProcessarLogHunt(ctx, s.db, corpo.Log)
	switch {
	case errors.Is(err, ErrSemSecaoLoot):
		escreverErro(w, http.StatusBadRequest, "Isso não parece um log do Hunt Analyser (não achei a seção de loot).")
		return
	case err != nil:
		log.Printf("erro ao processar log: %v", err)
		escreverErro(w, http.StatusInternalServerError, "Erro interno. Tente de novo.")
		return
	}

	resultado := []itemLog{}
	for _, it := range itens {
		resultado = append(resultado, itemLog(it)) // Item e itemLog têm os mesmos campos
	}
	escreverJSON(w, http.StatusOK, resultado)
}

// tratarErroBusca responde se houve erro e devolve true; devolve false se não houve.
func (s *servidor) tratarErroBusca(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrBuscaVazia), errors.Is(err, ErrBuscaLonga):
		// Erros "do usuário": a mensagem já é amigável.
		escreverErro(w, http.StatusBadRequest, err.Error())
	default:
		log.Printf("erro na busca: %v", err)
		escreverErro(w, http.StatusInternalServerError, "Erro interno. Tente de novo.")
	}
	return true
}

func escreverJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("erro ao escrever JSON: %v", err)
	}
}

func escreverErro(w http.ResponseWriter, status int, msg string) {
	escreverJSON(w, status, map[string]string{"erro": msg})
}

// --- Site (arquivos do front-end) -------------------------------------------

// servirSite entrega os arquivos de frontend/dist. O site pré-renderizado tem
// um HTML por página (buscar.html, analisador-de-log.html), então "/buscar"
// é atendido por "buscar.html". Endereços desconhecidos recebem a página
// 404.html (pré-renderizada) com status 404.
//
// Os arquivos são lidos por os.DirFS: o Go recusa, por conta própria, qualquer
// caminho que tente sair da pasta do site (como "../" ou, no Windows, "\").
func (s *servidor) servirSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}
	if _, err := os.Stat(s.site); err != nil {
		http.Error(w, "Site não compilado. Rode: cd frontend && npm run build", http.StatusServiceUnavailable)
		return
	}
	raiz := os.DirFS(s.site)

	// Nenhum endereço nosso tem barra invertida; no Windows ela separa pastas.
	if strings.ContainsRune(r.URL.Path, '\\') {
		http.Error(w, "endereço inválido", http.StatusBadRequest)
		return
	}

	// path.Clean resolve "." e ".." antes de qualquer acesso a arquivo.
	nome := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")

	var candidatos []string
	switch nome {
	case "":
		candidatos = []string{"index.html"}
	case "404", "404.html":
		// A página de erro só aparece como resposta a um endereço inexistente.
	default:
		candidatos = []string{nome, nome + ".html", nome + "/index.html"}
	}

	for _, arq := range candidatos {
		info, err := fs.Stat(raiz, arq)
		if err != nil || info.IsDir() {
			continue
		}
		// Arquivos em /assets têm hash no nome: podem ficar em cache por 1 ano.
		if strings.HasPrefix(nome, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			// Páginas e arquivos soltos: o navegador sempre confere se mudou,
			// assim um novo deploy aparece na hora.
			w.Header().Set("Cache-Control", "no-cache")
		}
		http.ServeFileFS(w, r, raiz, arq)
		return
	}

	pagina, err := fs.ReadFile(raiz, "404.html")
	if err != nil {
		pagina, _ = fs.ReadFile(raiz, "index.html") // site antigo, sem 404.html
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(pagina)
}
