package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Estes testes usam httptest: o servidor roda "de mentira", dentro do próprio
// teste, sem abrir porta nem precisar de banco de dados. Só testamos os
// caminhos que terminam ANTES de consultar o banco (validações, erros, arquivos).

// escrever cria um arquivo (e as pastas que faltarem) com o conteúdo dado.
func escrever(t *testing.T, caminho, conteudo string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(caminho), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(caminho, []byte(conteudo), 0o644); err != nil {
		t.Fatal(err)
	}
}

// siteDeTeste monta uma pasta parecida com frontend/dist. O arquivo
// "segredo.txt" fica FORA da pasta do site: ele nunca pode ser entregue.
func siteDeTeste(t *testing.T) string {
	t.Helper()
	raiz := t.TempDir()
	site := filepath.Join(raiz, "dist")
	escrever(t, filepath.Join(raiz, "segredo.txt"), "SEGREDO")
	escrever(t, filepath.Join(site, "index.html"), "<title>HOME</title>")
	escrever(t, filepath.Join(site, "buscar.html"), "<title>BUSCAR</title>")
	escrever(t, filepath.Join(site, "404.html"), "<title>PAGINA-404</title>")
	escrever(t, filepath.Join(site, "assets", "app-abc123.js"), "console.log(1)")
	return site
}

// pedir faz uma requisição ao handler e devolve a resposta.
func pedir(h http.Handler, metodo, alvo, corpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, alvo, strings.NewReader(corpo))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// mensagemDeErro lê o campo "erro" do JSON de resposta.
func mensagemDeErro(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var corpo map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("a resposta não é um JSON de erro: %q", rec.Body.String())
	}
	return corpo["erro"]
}

func TestSiteEntregaOsArquivos(t *testing.T) {
	h := (&servidor{site: siteDeTeste(t)}).rotas()

	casos := []struct {
		alvo, contem, cache string
	}{
		{"/", "HOME", "no-cache"},
		{"/buscar", "BUSCAR", "no-cache"},
		{"/assets/app-abc123.js", "console.log", "public, max-age=31536000, immutable"},
	}
	for _, c := range casos {
		rec := pedir(h, "GET", c.alvo, "")
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), c.contem) {
			t.Errorf("%s: status %d, corpo %q (esperava 200 com %q)", c.alvo, rec.Code, rec.Body.String(), c.contem)
		}
		if got := rec.Header().Get("Cache-Control"); got != c.cache {
			t.Errorf("%s: Cache-Control = %q, esperava %q", c.alvo, got, c.cache)
		}
	}
}

func TestSiteEnderecoInexistenteDevolvePagina404(t *testing.T) {
	h := (&servidor{site: siteDeTeste(t)}).rotas()

	// "/assets" é uma pasta: a listagem dela não pode aparecer.
	for _, alvo := range []string{"/nao-existe", "/404", "/404.html", "/assets", "/assets/nao-existe.js"} {
		rec := pedir(h, "GET", alvo, "")
		if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "PAGINA-404") {
			t.Errorf("%s: status %d, corpo %q (esperava 404 com a página 404.html)", alvo, rec.Code, rec.Body.String())
		}
	}
}

func TestSiteNaoSaiDaPastaDoSite(t *testing.T) {
	h := (&servidor{site: siteDeTeste(t)}).rotas()

	alvos := []string{
		"/../segredo.txt",
		"/..%2fsegredo.txt",
		"/%2e%2e/segredo.txt",
		"/assets/../../segredo.txt",
		"/assets/%2e%2e/%2e%2e/segredo.txt",
		"/....//segredo.txt",
	}
	for _, alvo := range alvos {
		rec := pedir(h, "GET", alvo, "")
		if strings.Contains(rec.Body.String(), "SEGREDO") {
			t.Errorf("%s: ENTREGOU um arquivo de fora da pasta do site!", alvo)
		}
		if rec.Code == http.StatusOK && !strings.Contains(rec.Body.String(), "HOME") {
			t.Errorf("%s: status 200 com corpo inesperado %q", alvo, rec.Body.String())
		}
	}

	// A barra invertida separa pastas no Windows: é recusada logo na entrada.
	req := httptest.NewRequest("GET", "/x", nil)
	req.URL.Path = `/..\segredo.txt`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || strings.Contains(rec.Body.String(), "SEGREDO") {
		t.Errorf("barra invertida: status %d, corpo %q (esperava 400)", rec.Code, rec.Body.String())
	}
}

func TestSiteSoAceitaLeitura(t *testing.T) {
	h := (&servidor{site: siteDeTeste(t)}).rotas()
	for _, metodo := range []string{"POST", "PUT", "DELETE", "PATCH"} {
		rec := pedir(h, metodo, "/buscar", "x")
		if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") == "" {
			t.Errorf("%s /buscar: status %d, esperava 405 com o cabeçalho Allow", metodo, rec.Code)
		}
	}
	if rec := pedir(h, "HEAD", "/buscar", ""); rec.Code != http.StatusOK {
		t.Errorf("HEAD /buscar: status %d, esperava 200", rec.Code)
	}
}

func TestSiteNaoCompilado(t *testing.T) {
	h := (&servidor{site: filepath.Join(t.TempDir(), "nao-existe")}).rotas()
	if rec := pedir(h, "GET", "/", ""); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("sem a pasta do site: status %d, esperava 503", rec.Code)
	}
}

func TestCabecalhosDeSeguranca(t *testing.T) {
	rec := pedir((&servidor{site: siteDeTeste(t)}).rotas(), "GET", "/", "")

	esperados := map[string]string{
		"X-Content-Type-Options":     "nosniff",
		"Referrer-Policy":            "strict-origin-when-cross-origin",
		"Cross-Origin-Opener-Policy": "same-origin",
	}
	for nome, valor := range esperados {
		if got := rec.Header().Get(nome); got != valor {
			t.Errorf("%s = %q, esperava %q", nome, got, valor)
		}
	}
	if csp := rec.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'self'") {
		t.Errorf("Content-Security-Policy = %q, deveria restringir ao próprio site", csp)
	}
	hsts := rec.Header().Get("Strict-Transport-Security")
	if !strings.Contains(hsts, "max-age=") || strings.Contains(hsts, "includeSubDomains") {
		t.Errorf("Strict-Transport-Security = %q (esperava max-age e SEM includeSubDomains)", hsts)
	}
}

func TestApiRecusaPedidosInvalidosAntesDoBanco(t *testing.T) {
	h := (&servidor{site: siteDeTeste(t)}).rotas() // db nil: qualquer acesso ao banco derrubaria o teste

	casos := []struct {
		nome, metodo, alvo, corpo string
		status                    int
		mensagem                  string // trecho esperado em "erro" (vazio = não confere)
	}{
		{"buscar com tipo inválido", "GET", "/api/buscar?tipo=x&q=a", "", 400, "tipo de busca inválido"},
		{"buscar com termo vazio", "GET", "/api/buscar?tipo=criatura&q=", "", 400, "ao menos um caractere"},
		{"buscar só com espaços", "GET", "/api/buscar?tipo=item&q=%20%20", "", 400, "ao menos um caractere"},
		{"buscar com termo longo", "GET", "/api/buscar?tipo=criatura&q=" + strings.Repeat("a", 101), "", 400, "longo demais"},
		{"buscar NPC com termo vazio", "GET", "/api/buscar?tipo=npc&q=", "", 400, "ao menos um caractere"},
		{"buscar NPC com termo longo", "GET", "/api/buscar?tipo=npc&q=" + strings.Repeat("a", 101), "", 400, "longo demais"},
		{"sugestões de NPC com termo longo", "GET", "/api/sugestoes?tipo=npc&q=" + strings.Repeat("a", 101), "", 400, "longo demais"},
		{"sugestões com tipo inválido", "GET", "/api/sugestoes?tipo=x&q=a", "", 400, "tipo de busca inválido"},
		{"sugestões com termo longo", "GET", "/api/sugestoes?tipo=item&q=" + strings.Repeat("a", 101), "", 400, "longo demais"},
		{"log com JSON malformado", "POST", "/api/log", "isto não é json", 400, "JSON"},
		{"log com corpo vazio", "POST", "/api/log", "", 400, "JSON"},
		{"log sem a seção de loot", "POST", "/api/log", `{"log":"oi"}`, 400, "seção de loot"},
		{"log maior que 64 KB", "POST", "/api/log", `{"log":"` + strings.Repeat("a", tamanhoMaxLog+1) + `"}`, 413, "64 KB"},
		{"corpo enorme", "POST", "/api/log", `{"log":"` + strings.Repeat("a", tamanhoMaxCorpoLog+10) + `"}`, 413, "64 KB"},
		{"método errado em /api/log", "GET", "/api/log", "", 405, "use POST"},
		{"método errado em /api/buscar", "POST", "/api/buscar", "", 405, "use GET"},
		{"endereço de API que não existe", "GET", "/api/nao-existe", "", 404, "não encontrado"},
	}
	for _, c := range casos {
		rec := pedir(h, c.metodo, c.alvo, c.corpo)
		if rec.Code != c.status {
			t.Errorf("%s: status %d, esperava %d (corpo %q)", c.nome, rec.Code, c.status, rec.Body.String())
			continue
		}
		if c.mensagem != "" && !strings.Contains(mensagemDeErro(t, rec), c.mensagem) {
			t.Errorf("%s: erro %q não contém %q", c.nome, mensagemDeErro(t, rec), c.mensagem)
		}
	}
}

func TestSugestoesComTermoVazioDevolveListaVazia(t *testing.T) {
	h := (&servidor{site: siteDeTeste(t)}).rotas()
	rec := pedir(h, "GET", "/api/sugestoes?tipo=criatura&q=", "")
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Errorf("status %d, corpo %q; esperava 200 com []", rec.Code, rec.Body.String())
	}
}

func TestLimiteDeRequisicoesSoValeParaAApi(t *testing.T) {
	h := (&servidor{site: siteDeTeste(t)}).rotas()

	// A rajada permitida é de 30: o 31º pedido seguido da mesma origem é barrado.
	bloqueadas := 0
	for i := 1; i <= 40; i++ {
		rec := pedir(h, "GET", "/api/buscar?tipo=x", "")
		if rec.Code == http.StatusTooManyRequests {
			bloqueadas++
			if i <= int(rajada) {
				t.Errorf("o pedido %d foi barrado, mas a rajada permite %d", i, int(rajada))
			}
			if rec.Header().Get("Retry-After") == "" {
				t.Error("a resposta 429 deveria ter o cabeçalho Retry-After")
			}
		}
	}
	if bloqueadas == 0 {
		t.Error("nenhum dos 40 pedidos seguidos foi barrado")
	}

	// As páginas do site nunca entram no limite.
	for i := 0; i < 60; i++ {
		if rec := pedir(h, "GET", "/", ""); rec.Code != http.StatusOK {
			t.Fatalf("a página inicial foi barrada no pedido %d (status %d)", i+1, rec.Code)
		}
	}
}

func TestIPDoCliente(t *testing.T) {
	casos := []struct {
		nome, confiarNoProxy, xff, esperado string
	}{
		{"sem proxy", "", "", "203.0.113.5"},
		{"sem proxy, cabeçalho forjado é ignorado", "", "1.1.1.1", "203.0.113.5"},
		{"com proxy usa o ÚLTIMO ip", "true", "1.1.1.1, 2.2.2.2", "2.2.2.2"},
		{"com proxy sem cabeçalho", "true", "", "203.0.113.5"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			t.Setenv("TRUST_PROXY", c.confiarNoProxy)
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = "203.0.113.5:4000"
			if c.xff != "" {
				req.Header.Set("X-Forwarded-For", c.xff)
			}
			if got := ipDoCliente(req); got != c.esperado {
				t.Errorf("ipDoCliente = %q, esperava %q", got, c.esperado)
			}
		})
	}
}

func TestSaudeComBancoForaDoAr(t *testing.T) {
	// Porta 1 não tem MySQL: a conexão é recusada na hora.
	db, err := sql.Open("mysql", "u:p@tcp(127.0.0.1:1)/x?timeout=1s")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := &servidor{db: db, site: siteDeTeste(t)}
	h := s.rotas()

	rec := pedir(h, "GET", "/healthz", "")
	if rec.Code != http.StatusServiceUnavailable || mensagemDeErro(t, rec) != "banco indisponível" {
		t.Fatalf("status %d, corpo %q; esperava 503 'banco indisponível'", rec.Code, rec.Body.String())
	}

	// O resultado é reaproveitado por alguns segundos: o segundo pedido
	// NÃO testa o banco de novo.
	primeiraVez := s.saudeEm
	if primeiraVez.IsZero() {
		t.Fatal("o resultado do teste de saúde deveria ter sido guardado")
	}
	time.Sleep(5 * time.Millisecond)
	if rec := pedir(h, "GET", "/healthz", ""); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("segundo pedido: status %d, esperava 503", rec.Code)
	}
	if !s.saudeEm.Equal(primeiraVez) {
		t.Error("o segundo pedido testou o banco de novo, mas deveria usar o resultado guardado")
	}
}
