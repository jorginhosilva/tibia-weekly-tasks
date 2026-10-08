# Tibia Weekly Tasks – Front-end

Vue 3 + Vite + Tailwind CSS v4, com **pré-renderização (SSG)** para SEO.

## Como rodar

```bash
cd frontend
npm install
npm run dev      # desenvolvimento (http://localhost:5173)
npm run build    # gera dist/ com um HTML por página
npm run preview  # testa o build
```

Requer Node 20.19+ ou 22.12+.

> **Não atualize o `@unhead/vue` para a versão 3.** O `vite-ssg` precisa da versão 2; com a 3 o build
> passa, mas o HTML sai **sem as tags de SEO** (title, description, canonical).

## Estrutura (leia nesta ordem)

| Pasta/arquivo | Papel |
|---|---|
| `src/config.js` | Nome, URL e descrição do site (**troque o domínio aqui**) |
| `src/main.js` | Ponto de entrada (cria app + router + SSG) |
| `src/router/routes.js` | Lista de páginas |
| `src/pages/` | Uma tela por rota |
| `src/components/` | Peças reutilizáveis (layout, home, search) |
| `src/composables/` | Lógica reutilizável (`useSeo`, `useAsyncAction`) |
| `src/services/tasksApi.js` | **Único** lugar que acessa dados (mock ↔ API Go) |
| `src/style.css` | Cores/tema Tailwind e classes `.btn`, `.card`, `.input` |

## SEO já incluído

HTML semântico (um `<h1>` por página, `<main>`, `<nav>`, `<ol>`), title/description/canonical/Open Graph por página,
JSON-LD (WebSite e FAQPage), `robots.txt`, `sitemap.xml`, HTML pré-renderizado e `noindex` na 404.

**Domínio:** `https://tibiatask.com.br` está em `src/config.js`, `public/robots.txt` e `public/sitemap.xml`
(se mudar, altere os três).

**Imagem de compartilhamento (`og:image`):** o PNG usado pelo site é `public/og-image.png` (1200x630).
O desenho editável é `design/og-image.svg` (abre no Figma, Canva ou Inkscape; ao mudar, exporte de novo em PNG
1200x630). As tags são geradas em `src/composables/useSeo.js` a partir de `SITE.ogImage` em `src/config.js`.

## Acessibilidade

Link "pular para o conteúdo", foco visível, `aria-live` nos resultados, rótulos em todos os campos,
`prefers-reduced-motion` respeitado, alvos de toque ≥ 44px.

## Conectar ao servidor Go

O servidor Go (`server.go`) expõe `GET /api/buscar`, `GET /api/sugestoes` (autocomplete) e `POST /api/log`,
e também entrega o site compilado de `frontend/dist`. Para publicar/testar tudo junto:

```bash
cd frontend && npm run build && cd ..
go run . web        # abre http://127.0.0.1:8080
```

Durante o desenvolvimento do front-end, deixe `go run . web` rodando e use `npm run dev`
(o Vite encaminha `/api` para a porta 8080). Para trabalhar sem o Go, crie `frontend/.env` com
`VITE_USE_MOCK=true` (usa poucos dados de exemplo).
