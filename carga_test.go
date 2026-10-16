package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
)

// --- Um "banco de mentira" -------------------------------------------------
//
// Em vez de um MySQL de verdade, o teste usa um driver que guarda tudo em
// memória e REGISTRA cada comando recebido. Assim dá para conferir quantas
// "viagens" a carga faz e se cada criatura foi ligada ao item certo.

type bancoFalso struct {
	mu        sync.Mutex
	comandos  []string
	maxArgs   int
	proximoID int64
	tabelas   map[string]map[string]int64 // tabela -> nome em minúsculas -> id
	grafias   map[string]map[string]string
	ligacoes  map[[2]int64]bool
}

func novoBancoFalso() *bancoFalso {
	return &bancoFalso{
		tabelas:  map[string]map[string]int64{"creatures": {}, "weekly_items": {}},
		grafias:  map[string]map[string]string{"creatures": {}, "weekly_items": {}},
		ligacoes: map[[2]int64]bool{},
	}
}

var (
	bancosFalsos   = map[string]*bancoFalso{}
	bancosFalsosMu sync.Mutex
)

func init() { sql.Register("bancofalso", driverFalso{}) }

type driverFalso struct{}

func (driverFalso) Open(nome string) (driver.Conn, error) {
	bancosFalsosMu.Lock()
	defer bancosFalsosMu.Unlock()
	b, ok := bancosFalsos[nome]
	if !ok {
		return nil, fmt.Errorf("banco falso %q não existe", nome)
	}
	return &conexaoFalsa{b: b}, nil
}

// abrirBancoFalso devolve um *sql.DB ligado a um banco de mentira novo.
func abrirBancoFalso(t *testing.T) (*sql.DB, *bancoFalso) {
	t.Helper()
	b := novoBancoFalso()
	nome := t.Name()
	bancosFalsosMu.Lock()
	bancosFalsos[nome] = b
	bancosFalsosMu.Unlock()

	db, err := sql.Open("bancofalso", nome)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1) // uma conexão só: a transação e as consultas veem o mesmo estado
	t.Cleanup(func() { db.Close() })
	return db, b
}

type conexaoFalsa struct{ b *bancoFalso }

func (c *conexaoFalsa) Prepare(q string) (driver.Stmt, error) {
	return &comandoFalso{b: c.b, q: q}, nil
}
func (c *conexaoFalsa) Close() error              { return nil }
func (c *conexaoFalsa) Begin() (driver.Tx, error) { return txFalsa{}, nil }

type txFalsa struct{}

func (txFalsa) Commit() error   { return nil }
func (txFalsa) Rollback() error { return nil }

type comandoFalso struct {
	b *bancoFalso
	q string
}

func (s *comandoFalso) Close() error  { return nil }
func (s *comandoFalso) NumInput() int { return -1 }

func (s *comandoFalso) Exec(args []driver.Value) (driver.Result, error) {
	b := s.b
	b.mu.Lock()
	defer b.mu.Unlock()
	b.comandos = append(b.comandos, s.q)
	if len(args) > b.maxArgs {
		b.maxArgs = len(args)
	}

	switch {
	case strings.HasPrefix(s.q, "DELETE FROM "):
		tabela := strings.TrimPrefix(s.q, "DELETE FROM ")
		switch tabela {
		case "creature_drops":
			b.ligacoes = map[[2]int64]bool{}
		case "creatures", "weekly_items":
			b.tabelas[tabela] = map[string]int64{}
			b.grafias[tabela] = map[string]string{}
		}
		return driver.RowsAffected(0), nil

	case strings.HasPrefix(s.q, "INSERT INTO creatures (name)"), strings.HasPrefix(s.q, "INSERT INTO weekly_items (name)"):
		tabela := "creatures"
		if strings.HasPrefix(s.q, "INSERT INTO weekly_items") {
			tabela = "weekly_items"
		}
		for _, a := range args {
			nome := a.(string)
			chave := strings.ToLower(nome)
			if _, existe := b.tabelas[tabela][chave]; existe {
				continue // ON DUPLICATE KEY UPDATE: não muda nada
			}
			b.proximoID++
			b.tabelas[tabela][chave] = b.proximoID
			b.grafias[tabela][chave] = nome
		}
		return driver.RowsAffected(int64(len(args))), nil

	case strings.HasPrefix(s.q, "INSERT INTO creature_drops"):
		for i := 0; i < len(args); i += 2 {
			b.ligacoes[[2]int64{args[i].(int64), args[i+1].(int64)}] = true
		}
		return driver.RowsAffected(int64(len(args) / 2)), nil
	}
	return nil, fmt.Errorf("comando inesperado no banco falso: %q", s.q)
}

func (s *comandoFalso) Query(args []driver.Value) (driver.Rows, error) {
	b := s.b
	b.mu.Lock()
	defer b.mu.Unlock()
	b.comandos = append(b.comandos, s.q)

	var tabela string
	switch s.q {
	case "SELECT id, name FROM creatures":
		tabela = "creatures"
	case "SELECT id, name FROM weekly_items":
		tabela = "weekly_items"
	default:
		return nil, fmt.Errorf("consulta inesperada no banco falso: %q", s.q)
	}
	var linhas [][]driver.Value
	for chave, id := range b.tabelas[tabela] {
		linhas = append(linhas, []driver.Value{id, b.grafias[tabela][chave]})
	}
	return &linhasFalsas{linhas: linhas}, nil
}

type linhasFalsas struct {
	linhas [][]driver.Value
	pos    int
}

func (r *linhasFalsas) Columns() []string { return []string{"id", "name"} }
func (r *linhasFalsas) Close() error      { return nil }
func (r *linhasFalsas) Next(dest []driver.Value) error {
	if r.pos >= len(r.linhas) {
		return io.EOF
	}
	copy(dest, r.linhas[r.pos])
	r.pos++
	return nil
}

// --- Testes ------------------------------------------------------------------

// O teste principal: gravar a lista REAL do populate.go no banco falso e conferir
// que tudo entrou certo, com poucas viagens.
func TestGravarDropsEmLotesGravaOQueEstaNoArquivo(t *testing.T) {
	db, banco := abrirBancoFalso(t)

	if err := gravarDrops(context.Background(), db, true); err != nil {
		t.Fatalf("gravarDrops: %v", err)
	}

	// Conta, de forma independente, o que deveria existir.
	criaturas, itens := map[string]bool{}, map[string]bool{}
	for _, d := range dropsIniciais {
		criaturas[strings.ToLower(d.Criatura)] = true
		itens[strings.ToLower(d.Item)] = true
	}
	if got := len(banco.tabelas["creatures"]); got != len(criaturas) {
		t.Errorf("criaturas gravadas = %d, esperava %d", got, len(criaturas))
	}
	if got := len(banco.tabelas["weekly_items"]); got != len(itens) {
		t.Errorf("itens gravados = %d, esperava %d", got, len(itens))
	}
	if got := len(banco.ligacoes); got != len(dropsIniciais) {
		t.Errorf("ligações gravadas = %d, esperava %d", got, len(dropsIniciais))
	}

	// Cada par do arquivo precisa estar ligado aos ids CERTOS.
	for _, d := range dropsIniciais {
		c := banco.tabelas["creatures"][strings.ToLower(d.Criatura)]
		i := banco.tabelas["weekly_items"][strings.ToLower(d.Item)]
		if !banco.ligacoes[[2]int64{c, i}] {
			t.Fatalf("o par %q -> %q não foi ligado corretamente (ids %d e %d)", d.Criatura, d.Item, c, i)
		}
	}

	// O objetivo da mudança: poucas "viagens" ao banco (antes eram mais de 7.000).
	if viagens := len(banco.comandos); viagens > 30 {
		t.Errorf("a carga fez %d comandos; esperava no máximo 30:\n%s", viagens, strings.Join(banco.comandos, "\n"))
	}
	// E nenhum comando pode passar do limite de parâmetros do MySQL.
	if banco.maxArgs > 65535 {
		t.Errorf("um comando usou %d parâmetros; o MySQL aceita até 65535", banco.maxArgs)
	}
}

func TestRecriarApagaNaOrdemCertaAntesDeGravar(t *testing.T) {
	db, banco := abrirBancoFalso(t)
	if err := gravarDrops(context.Background(), db, true); err != nil {
		t.Fatal(err)
	}
	// As ligações dependem das outras duas tabelas, então são apagadas primeiro.
	esperado := []string{"DELETE FROM creature_drops", "DELETE FROM creatures", "DELETE FROM weekly_items"}
	for i, e := range esperado {
		if banco.comandos[i] != e {
			t.Errorf("comando %d = %q, esperava %q", i+1, banco.comandos[i], e)
		}
	}
}

func TestInserirSemApagarNaoRodaDelete(t *testing.T) {
	db, banco := abrirBancoFalso(t)
	if err := gravarDrops(context.Background(), db, false); err != nil {
		t.Fatal(err)
	}
	for _, c := range banco.comandos {
		if strings.HasPrefix(c, "DELETE") {
			t.Errorf("o modo que só acrescenta não deveria apagar nada, mas rodou %q", c)
		}
	}
}

func TestRodarDeNovoNaoDuplicaNada(t *testing.T) {
	db, banco := abrirBancoFalso(t)
	for i := 0; i < 2; i++ {
		if err := gravarDrops(context.Background(), db, false); err != nil {
			t.Fatal(err)
		}
	}
	if got := len(banco.ligacoes); got != len(dropsIniciais) {
		t.Errorf("depois de gravar duas vezes há %d ligações, esperava %d", got, len(dropsIniciais))
	}
}

func TestLotesRespeitamOTamanhoMaximo(t *testing.T) {
	// 1.203 criaturas e 1.203 itens distintos, ligados um a um: com lotes de 500,
	// são 3 comandos para as criaturas, 3 para os itens e 3 para as ligações.
	const n = 1203
	var drops []dropInicial
	for i := 0; i < n; i++ {
		drops = append(drops, dropInicial{fmt.Sprintf("Criatura %04d", i), fmt.Sprintf("Item %04d", i)})
	}
	db, banco := abrirBancoFalso(t)
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := gravarEmLotes(context.Background(), tx, drops); err != nil {
		t.Fatal(err)
	}

	var insCriaturas, insItens, insLigacoes int
	for _, c := range banco.comandos {
		switch {
		case strings.HasPrefix(c, "INSERT INTO creatures"):
			insCriaturas++
		case strings.HasPrefix(c, "INSERT INTO weekly_items"):
			insItens++
		case strings.HasPrefix(c, "INSERT INTO creature_drops"):
			insLigacoes++
		}
	}
	if insCriaturas != 3 || insItens != 3 || insLigacoes != 3 {
		t.Errorf("comandos de criaturas/itens/ligações = %d/%d/%d, esperava 3/3/3", insCriaturas, insItens, insLigacoes)
	}
	if banco.maxArgs != tamanhoLote*2 {
		t.Errorf("o maior comando usou %d parâmetros, esperava %d (lote cheio de ligações)", banco.maxArgs, tamanhoLote*2)
	}
	if got := len(banco.ligacoes); got != n {
		t.Errorf("ligações gravadas = %d, esperava %d", got, n)
	}
}

func TestMarcadores(t *testing.T) {
	casos := []struct {
		linhas, colunas int
		esperado        string
	}{
		{1, 1, "(?)"},
		{3, 1, "(?),(?),(?)"},
		{2, 2, "(?,?),(?,?)"},
	}
	for _, c := range casos {
		if got := marcadores(c.linhas, c.colunas); got != c.esperado {
			t.Errorf("marcadores(%d, %d) = %q, esperava %q", c.linhas, c.colunas, got, c.esperado)
		}
	}
}

func TestNomesDistintos(t *testing.T) {
	drops := []dropInicial{
		{"Wyrm", "Wyrm Scale"},
		{"wyrm", "Wyrm Heart"},   // mesma criatura, outra grafia
		{"Dragon", "wyrm scale"}, // mesmo item, outra grafia
		{"Dragon", "Dragon Claw"},
	}
	criaturas, itens := nomesDistintos(drops)

	if got := strings.Join(criaturas, ","); got != "Wyrm,Dragon" {
		t.Errorf("criaturas = %q, esperava %q (primeira grafia, na ordem de chegada)", got, "Wyrm,Dragon")
	}
	if got := strings.Join(itens, ","); got != "Wyrm Scale,Wyrm Heart,Dragon Claw" {
		t.Errorf("itens = %q, esperava %q", got, "Wyrm Scale,Wyrm Heart,Dragon Claw")
	}
}
