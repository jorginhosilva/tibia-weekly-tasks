package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// logDeTeste é um exemplo de log do Hunt Analyser.
const logDeTeste = `Session data: From 2026-09-24, 22:37:41 to 2026-09-24, 22:52:05
Session: 00:14h
Raw XP Gain: 417,641
XP Gain: 626,461
Raw XP/h: 1,737,705
XP/h: 2,606,555
Loot: 251,888
Supplies: 45,464
Balance: 206,424
Damage: 691,136
Damage/h: 3,397,933
Healing: 151,679
Healing/h: 745,721
Killed Monsters:
  37x bashmu
  52x juvenile bashmu
Looted Items:
  1x a terra amulet
  2x a glacier amulet
  1x a small diamond
  707x a platinum coin
  2x a violet gem
  1x a yellow gem
  1x a red gem
  1x a war hammer
  7x a guardian halberd
  1x a crystal mace
  3x a diamond sceptre
  1x an amber staff
  1x a chaos mace
  1x a dragonbone staff
  11x a great spirit potion
  15x an ultimate health potion
  4x a sacred tree amulet
  1x a blue crystal shard
  1x a violet crystal shard
  2x a green crystal shard
  2x a cyan crystal fragment
  1x a pair of iron fists
  1x a collar of red plasma
  1x a bashmu fang
  4x a bashmu tongue
  4x a bashmu feather`

// main só chama run() e decide o que fazer se ela falhar.
// Motivo: log.Fatal encerra o programa NA HORA e pularia o "defer db.Close()".
// Dentro de run(), os defer funcionam normalmente.
func main() {
	if err := run(); err != nil {
		log.Fatalf("erro: %v", err)
	}
}

// textoUso explica os modos do programa. Aparece em "go run . ajuda" e quando
// o argumento digitado não existe.
const textoUso = `Uso:
  go run .                      carrega os dados (só acrescenta) e mostra uma demonstração
  go run . web                  sobe o site em http://127.0.0.1:8080
  go run . recriar --confirmo   APAGA as tabelas e recarrega tudo do populate.go
  go run . ajuda                mostra esta mensagem
`

// modoExecucao é o que o programa vai fazer, escolhido pelo argumento.
type modoExecucao string

const (
	modoDemo    modoExecucao = "demo" // sem argumento
	modoWeb     modoExecucao = "web"
	modoRecriar modoExecucao = "recriar"
	modoAjuda   modoExecucao = "ajuda"
)

// interpretarArgumentos converte os argumentos da linha de comando em um modo.
// Um argumento desconhecido (por exemplo, "wbe" no lugar de "web") é recusado
// em vez de cair no modo que grava no banco.
// O segundo valor diz se o usuário passou --confirmo (exigido pelo "recriar").
func interpretarArgumentos(args []string) (modoExecucao, bool, error) {
	if len(args) == 0 {
		return modoDemo, false, nil
	}

	modo, resto := args[0], args[1:]
	switch modo {
	case "web":
		if len(resto) > 0 {
			return "", false, fmt.Errorf("o modo web não aceita argumentos extras (%q)\n\n%s", resto, textoUso)
		}
		return modoWeb, false, nil
	case "recriar":
		switch {
		case len(resto) == 0:
			return modoRecriar, false, nil
		case len(resto) == 1 && resto[0] == "--confirmo":
			return modoRecriar, true, nil
		}
		return "", false, fmt.Errorf("argumento desconhecido para recriar: %q\n\n%s", resto, textoUso)
	case "ajuda", "help", "-h", "--help":
		return modoAjuda, false, nil
	}
	return "", false, fmt.Errorf("modo desconhecido: %q\n\n%s", modo, textoUso)
}

func run() error {
	modo, confirmado, err := interpretarArgumentos(os.Args[1:])
	if err != nil {
		return err
	}
	if modo == modoAjuda {
		fmt.Print(textoUso)
		return nil
	}

	cfg, err := carregarConfigBanco()
	if err != nil {
		return err
	}

	// "go run . web" sobe o site local.
	if modo == modoWeb {
		return rodarSite(cfg)
	}

	// O recriar APAGA dados: antes de qualquer conexão, mostramos o alvo e
	// exigimos --confirmo, para ninguém apagar o banco errado por engano.
	if modo == modoRecriar && !confirmado {
		return fmt.Errorf("o modo recriar APAGA todos os dados do banco %q em %s e os recarrega do populate.go.\n"+
			"Se é isso mesmo, rode:  go run . recriar --confirmo", cfg.DBName, cfg.Addr)
	}

	// O contexto dá um limite de tempo para tudo que fala com o banco.
	ctx, cancelar := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancelar()

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return fmt.Errorf("abrindo a conexão com o banco: %w", err)
	}
	defer db.Close()

	// sql.Open NÃO conecta de verdade. O Ping é que testa se o banco responde.
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("conectando ao banco (confira o .env e se o MySQL está ligado): %w", err)
	}
	fmt.Println("DB Conectado!")

	if err := CriarTabelas(ctx, db); err != nil {
		return err
	}
	fmt.Println("Tabelas prontas.")

	// "go run . recriar" deixa o banco IGUAL ao populate.go (apaga o que não
	// está no arquivo). Use depois de corrigir qualquer nome no arquivo.
	if modo == modoRecriar {
		fmt.Printf("Recriando o banco %q em %s...\n", cfg.DBName, cfg.Addr)
		if err := RecriarDadosTibia(ctx, db); err != nil {
			return err
		}
		return imprimirContagens(ctx, db)
	}

	if err := InserirDadosTibia(ctx, db); err != nil {
		return err
	}
	fmt.Printf("%d drops cadastrados/conferidos.\n", len(dropsIniciais))

	itens, err := ProcessarLogHunt(ctx, db, logDeTeste)
	switch {
	case errors.Is(err, ErrSemSecaoLoot):
		// Erro "esperado": o texto colado não parece um log. Avisamos sem alarde.
		fmt.Println("O texto informado não parece um log do Hunt Analyser (sem seção de loot).")
	case err != nil:
		return err
	default:
		imprimirResultado(itens)
	}

	return demonstrarBuscas(ctx, db)
}

// rodarSite abre o banco e sobe o servidor web em http://127.0.0.1:8080.
// 127.0.0.1 significa "só este computador": ninguém na rede consegue acessar.
// Não cria tabelas nem carrega dados; isso é feito pelo modo antigo (go run .).
func rodarSite(cfg *mysql.Config) error {
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return fmt.Errorf("abrindo a conexão com o banco: %w", err)
	}
	defer db.Close()

	ctx, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("conectando ao banco: %w", err)
	}

	// Em hospedagens (Railway, Render, Fly...) a plataforma define a porta na
	// variável PORT e o site precisa escutar em todas as interfaces.
	addr := "127.0.0.1:8080"
	if porta := os.Getenv("PORT"); porta != "" {
		addr = ":" + porta
	}
	// Limita as conexões abertas com o banco, para um pico de acessos não
	// esgotar o limite do MySQL da hospedagem.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	srv := NovoServidorHTTP(db, addr)

	// Encerramento limpo: ao receber Ctrl+C ou o SIGTERM da hospedagem (que o
	// envia a cada novo deploy), terminamos as requisições em andamento em vez
	// de cortá-las no meio.
	parar, pararSinal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer pararSinal()
	erroServidor := make(chan error, 1)
	go func() { erroServidor <- srv.ListenAndServe() }()

	fmt.Printf("Site no ar: %s  (Ctrl+C para parar)\n", addr)
	select {
	case err := <-erroServidor:
		return err
	case <-parar.Done():
		fmt.Println("Encerrando...")
		fim, cancelarFim := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelarFim()
		return srv.Shutdown(fim)
	}
}

// demonstrarBuscas mostra as buscas funcionando. Quando o site existir, os
// termos virão de campos de busca na página, e não de textos fixos aqui.
func demonstrarBuscas(ctx context.Context, db *sql.DB) error {
	const criaturaDeTeste = "bashmu"
	const itemDeTeste = "chaos mace"

	porCriatura, err := BuscarPorCriatura(ctx, db, criaturaDeTeste)
	if err != nil {
		return err
	}
	fmt.Printf("\n🔎 Criaturas com %q e seus itens de weekly task:\n", criaturaDeTeste)
	if len(porCriatura) == 0 {
		fmt.Println("  nenhuma criatura encontrada")
	}
	for _, c := range porCriatura {
		fmt.Printf("  %s: %s\n", c.Criatura, strings.Join(c.Itens, ", "))
	}

	porItem, err := BuscarPorItem(ctx, db, itemDeTeste)
	if err != nil {
		return err
	}
	fmt.Printf("\n🔎 Itens com %q e as criaturas que os dropam:\n", itemDeTeste)
	if len(porItem) == 0 {
		fmt.Println("  nenhum item encontrado")
	}
	for _, i := range porItem {
		fmt.Printf("  %s: %s\n", i.Item, strings.Join(i.Criaturas, ", "))
	}

	return nil
}

// carregarConfigBanco lê o .env (ou as variáveis de ambiente) e monta a
// configuração de conexão. Usar mysql.Config em vez de montar o texto da
// conexão à mão evita problemas com senhas que têm caracteres como @ ou /.
func carregarConfigBanco() (*mysql.Config, error) {
	// Se o arquivo .env não existir, tudo bem: as variáveis podem já estar
	// definidas no sistema. Qualquer OUTRO erro de leitura é um problema real.
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("lendo o arquivo .env: %w", err)
	}

	usuario := os.Getenv("DB_USER")
	host := os.Getenv("DB_HOST")
	porta := os.Getenv("DB_PORT")
	banco := os.Getenv("DB_NAME")

	// Uma lista (e não um map) mantém a ordem, e assim a mensagem cita sempre
	// TODAS as variáveis que faltam, na mesma ordem.
	obrigatorias := []struct{ nome, valor string }{
		{"DB_USER", usuario},
		{"DB_HOST", host},
		{"DB_PORT", porta},
		{"DB_NAME", banco},
	}
	var faltando []string
	for _, v := range obrigatorias {
		if v.valor == "" {
			faltando = append(faltando, v.nome)
		}
	}
	if len(faltando) > 0 {
		return nil, fmt.Errorf("variáveis não definidas no .env: %s", strings.Join(faltando, ", "))
	}

	cfg := mysql.NewConfig()
	cfg.User = usuario
	cfg.Passwd = os.Getenv("DB_PASSWORD")
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(host, porta)
	cfg.DBName = banco
	cfg.ParseTime = true
	// Bancos na nuvem exigem conexão criptografada. Em produção, defina DB_TLS=true.
	if os.Getenv("DB_TLS") == "true" {
		cfg.TLS = &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	}
	cfg.AllowNativePasswords = true

	return cfg, nil
}

// imprimirContagens mostra quantas linhas o banco tem agora, para conferir
// com o populate.go depois de recriar.
func imprimirContagens(ctx context.Context, db *sql.DB) error {
	var criaturas, itens, ligacoes int
	err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM creatures),
		(SELECT COUNT(*) FROM weekly_items),
		(SELECT COUNT(*) FROM creature_drops)`).Scan(&criaturas, &itens, &ligacoes)
	if err != nil {
		return fmt.Errorf("contando os registros: %w", err)
	}
	fmt.Printf("Banco recriado a partir do populate.go: %d criaturas, %d itens, %d ligações.\n", criaturas, itens, ligacoes)
	return nil
}

// imprimirResultado é a única parte que "conversa" com o terminal.
func imprimirResultado(itens []Item) {
	fmt.Println("\n📋 ITENS DE WEEKLY TASK ENCONTRADOS NA HUNT:")
	fmt.Println("=======================================================")

	if len(itens) == 0 {
		fmt.Println("Nenhum item de Weekly Task foi identificado neste log de hunt.")
		return
	}

	for _, item := range itens {
		fmt.Printf("🚨 ITEM DE TASK ENCONTRADO: %dx %s\n", item.Quantidade, item.Nome)
	}
}
