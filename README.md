# Tibia Weekly Tasks

Site gratuito para jogadores brasileiros de **Tibia** descobrirem, em segundos, **quais criaturas dropam os itens da Weekly Task** e **quais itens de task já coletaram** na caçada, colando o log do Hunt Analyser.

> Endereço: **https://tibiatask.com.br** (em breve)

## O que o site faz

- **Busca por criatura:** digite um monstro e veja os itens de task que ele dropa.
- **Busca por item:** digite um item e veja quais criaturas o dropam.
- **Autocomplete:** sugestões a partir de 2 letras, navegáveis por teclado e compatíveis com leitores de tela.
- **Analisador de log:** cole o log do Hunt Analyser e receba só os itens que importam para a task, com as quantidades somadas.
- **Banco com 872 criaturas, 471 itens e 2.331 ligações** (criatura → item), mantido em um único arquivo versionado.

## Tecnologias

| Camada | Escolha |
|---|---|
| Back-end | Go 1.26 (somente biblioteca padrão + driver MySQL + godotenv) |
| Banco de dados | MySQL 9 |
| Front-end | Vue 3, Vue Router, Vite, Tailwind CSS v4 |
| SEO | Pré-renderização (SSG) com `vite-ssg`: um HTML pronto por página, com title, description, canonical, Open Graph e dados estruturados (JSON-LD) |
| Hospedagem | Docker (uma imagem com o servidor Go e o site compilado) |

## Como rodar localmente

Requisitos: **Go 1.26+**, **Node 20.19+ (ou 22.12+)** e um **MySQL** (abaixo, um exemplo com Docker).

```bash
# 1) MySQL de desenvolvimento (troque a senha)
docker run --name tibia-mysql -e MYSQL_ROOT_PASSWORD=TROQUE_ESTA_SENHA \
  -e MYSQL_DATABASE=tibia_tasks_db -p 3306:3306 -d mysql:9

# 2) Configuração: copie o exemplo e preencha usuário e senha
cp .env.example .env

# 3) Cria as tabelas e carrega os dados do populate.go
go run . recriar --confirmo

# 4) Compila o site e sobe tudo em http://127.0.0.1:8080
cd frontend && npm install && npm run build && cd ..
go run . web
```

Para desenvolver o front-end com recarregamento automático, deixe `go run . web` rodando e, em outro terminal, use `cd frontend && npm run dev` (o Vite encaminha `/api` para a porta 8080).

### Comandos do programa

| Comando | O que faz |
|---|---|
| `go run .` | Carrega os dados (**só acrescenta**), recarrega os preços dos NPCs e mostra uma demonstração |
| `go run . web` | Sobe o site em `http://127.0.0.1:8080` |
| `go run . recriar --confirmo` | **Apaga** as tabelas e recarrega tudo do `populate.go` e de `dados/npcs` (atômico: se falhar, nada muda) |
| `go run . ajuda` | Mostra o uso |

Sem `--confirmo`, o `recriar` só mostra qual banco seria apagado e para.

## Como manter os dados

Todos os dados ficam em [populate.go](populate.go), uma linha por ligação: `{"Criatura", "Item"}`, em **ordem alfabética por criatura**.

1. Edite o `populate.go`.
2. Rode `go test ./...`. Os testes conferem os dados reais e **recusam** nome vazio, espaço sobrando, ligação repetida, mesmo nome com maiúsculas diferentes, criatura e item trocados e nomes que diferem em uma letra (erros de digitação).
3. Rode `go run . recriar --confirmo` para deixar o banco **idêntico** ao arquivo.

A carga também aplica essa validação: dados com erro nunca chegam ao banco.

### Preços dos NPCs

Os preços ficam em [dados/npcs](dados/npcs), **um arquivo por NPC**. O nome do arquivo vira o nome do NPC (`rashid.csv` → Rashid, `esrik_pompan.csv` → Esrik Pompan):

```
item;preco
Chaos Mace;9000
```

- Preço só com números, sem ponto nem vírgula. Um erro de **formato** (preço inválido, linha sem `;`, item repetido) trava a carga e nada é gravado.
- O nome do item é guardado como texto. Um nome que não está na lista da weekly **é gravado mesmo assim** e aparece como aviso no terminal, para ser corrigido depois.
- `go run .` e `go run . recriar --confirmo` deixam os preços do banco **iguais** aos arquivos da pasta. Se a pasta não existir ou estiver vazia, os preços do banco não são alterados.
- `go test ./...` confere o formato dos arquivos reais.

## API

| Método e caminho | Resposta |
|---|---|
| `GET /api/buscar?tipo=criatura\|item\|npc&q=texto` | `[{ "nome": "...", "relacionados": ["..."], "precos": [{ "nome": "...", "preco": 9000 }] }]` |
| `GET /api/sugestoes?tipo=criatura\|item\|npc&q=texto` | `["nome", ...]` (até 12) |
| `POST /api/log` com `{ "log": "texto" }` | `[{ "nome": "...", "quantidade": 3 }]` |
| `GET /healthz` | `{ "status": "ok" }` quando o banco responde |

`precos` só aparece quando há preços: na busca por NPC, são os itens e quanto ele paga; na busca por item, os NPCs que compram o item.

A busca `tipo=npc` aceita o nome do NPC **ou** de um item: o NPC cujo nome contém o termo vem com todos os itens; os outros vêm só com os itens que contêm o termo (ex.: `banana sash` traz só esse item, no NPC que o compra). As sugestões desse tipo trazem nomes de NPC e de itens.

Erros voltam com status 4xx/5xx e `{ "erro": "mensagem amigável" }`.

## Estrutura

```
.
├── main.go          modos do programa (web, recriar, demonstração) e leitura do .env
├── server.go        rotas HTTP, API JSON e entrega do site compilado
├── ratelimit.go     limite de requisições por visitante
├── search.go        buscas e sugestões (SQL parametrizado)
├── parser.go        leitura do log do Hunt Analyser
├── repository.go    gravação de criaturas, itens e ligações
├── database.go      criação das tabelas
├── populate.go      os dados (criatura → item)
├── validacao.go     conferência dos dados antes de gravar
├── npcs.go          leitura e carga dos preços dos NPCs
├── dados/npcs/      preços dos NPCs, um .csv por NPC (preenchidos à mão)
├── *_test.go        testes automáticos
├── Dockerfile       imagem de produção
└── frontend/        site em Vue (veja frontend/README.md)
```

## Segurança

- **SQL injection:** todas as consultas usam parâmetros; o texto digitado nunca é colado no SQL. Os curingas do `LIKE` (`%` e `_`) são escapados.
- **Limite de requisições:** 10 por segundo por visitante (rajada de 30) na API, com resposta 429.
- **Cabeçalhos:** Content-Security-Policy restrita ao próprio site, HSTS, `nosniff`, Referrer-Policy e Permissions-Policy.
- **Limites de tamanho:** busca com no máximo 100 caracteres e log com no máximo 64 KB.
- **Arquivos:** o site é lido por `os.DirFS`, que recusa caminhos fora da pasta (`../`, `\`), e há testes para isso.
- **Segredos:** o `.env` fica fora do git (`.env.example` mostra o formato). Em produção, use um usuário do MySQL **somente leitura** para o site.
- **Privacidade:** o log colado é analisado na hora e **não é armazenado**.

## Testes

```bash
go vet ./...
go test ./...
cd frontend && npm run build
```

## Publicação (Docker)

```bash
docker build -t tibiatask .
# DB_HOST=127.0.0.1 não funciona dentro do contêiner: use host.docker.internal (ou o endereço do banco)
docker run -p 8080:8080 --env-file .env -e DB_HOST=host.docker.internal -e PORT=8080 tibiatask
```

Variáveis de ambiente do servidor:

| Variável | Para quê |
|---|---|
| `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME` | Conexão com o MySQL |
| `PORT` | Porta do servidor (as hospedagens definem sozinhas) |
| `TRUST_PROXY=true` | Atrás do proxy da hospedagem: usa o IP do visitante no limite de requisições |
| `DB_TLS=true` | Conexão criptografada com o banco, quando o provedor a exigir |
| `SITE_DIR` | Pasta do site compilado (padrão `frontend/dist`) |

Configure o health check da hospedagem para `/healthz`.

## Como foi construído

O projeto nasceu como um programa em Go (leitura do log e banco de dados). O **Claude Code** (Anthropic) foi usado para construir o **front-end** (Vue + Tailwind, SEO e acessibilidade), a **API HTTP** (endpoints, autocomplete, limite de requisições e cabeçalhos de segurança), a **validação dos dados** e para **revisar o código**. As decisões, os requisitos, a conferência no navegador contra o banco real e a aprovação de cada mudança foram do autor.

## Dados e marcas

Os nomes das criaturas e dos itens, e a relação de quais criaturas dropam cada item da Weekly Task, foram reunidos
manualmente a partir de consultas a estas fontes:

- **[Tibia Wiki](https://www.tibiawiki.com.br)** (tibiawiki.com.br)
- **[Tibiopedia.pl](https://tibiopedia.pl)**

O projeto guarda apenas **nomes e relações** (fatos do jogo). **Nenhum texto, imagem ou arte** dessas wikis é copiado
ou exibido. Os dados são conteúdo do jogo Tibia e as wikis são trabalho de suas comunidades: este site é apenas uma
ferramenta de consulta e não substitui as fontes originais. Para informações completas sobre criaturas e itens,
consulte as próprias wikis.

> Se você mantém alguma dessas fontes e quer ajustar o crédito ou pedir a remoção de algo, abra uma *issue* neste
> repositório.

Tibia é marca registrada da **CipSoft GmbH**. Este é um projeto de fãs, sem vínculo com a empresa nem com as wikis acima.
