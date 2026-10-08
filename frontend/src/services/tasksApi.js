import { MOCK_DROPS } from '@/data/mockData'

/**
 * Camada de acesso a dados. As páginas NUNCA chamam `fetch` diretamente:
 * só usam estas funções. Assim, trocar a origem dos dados exige mexer em UM arquivo.
 *
 * Endpoints do servidor Go (server.go):
 *   GET  /api/buscar?tipo=criatura|item&q=texto     -> [{ nome, relacionados: [string] }]
 *   GET  /api/sugestoes?tipo=criatura|item&q=texto  -> [string]   (autocomplete)
 *   POST /api/log   { log: "texto" }                -> [{ nome, quantidade }]
 *   Erros: status 4xx/5xx com { erro: "mensagem amigável" }
 *
 * Por padrão usa a API real. Para trabalhar SEM o servidor Go, crie um
 * arquivo .env com VITE_USE_MOCK=true (usa poucos dados de exemplo).
 */
const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true'

/** Chama a API e devolve o JSON; em caso de erro, lança a mensagem amigável do servidor. */
async function callApi(url, options) {
  let res
  try {
    res = await fetch(url, options)
  } catch (e) {
    if (e.name === 'AbortError') throw e // cancelamento é esperado (autocomplete)
    throw new Error('Não foi possível conectar ao servidor. Tente novamente.')
  }
  const body = await res.json().catch(() => null)
  if (!res.ok) throw new Error(body?.erro || 'Algo deu errado. Tente novamente.')
  return body
}

// --- Mock (só para desenvolvimento sem backend) -----------------------------
const wait = (ms) => new Promise((r) => setTimeout(r, ms))
const norm = (s) => s.normalize('NFD').replace(/\p{M}/gu, '').toLowerCase()
const campo = (tipo) => (tipo === 'criatura' ? 'criatura' : 'item')

function group(drops, by, other) {
  const map = new Map()
  for (const d of drops) {
    if (!map.has(d[by])) map.set(d[by], [])
    map.get(d[by]).push(d[other])
  }
  return [...map].map(([nome, relacionados]) => ({ nome, relacionados }))
}

// --- Funções públicas --------------------------------------------------------

/** Busca por criatura (retorna itens) ou por item (retorna criaturas). */
export async function search(tipo, termo) {
  if (!USE_MOCK) {
    const qs = new URLSearchParams({ tipo, q: termo })
    return callApi(`/api/buscar?${qs}`)
  }
  await wait(200)
  const by = campo(tipo)
  const other = by === 'criatura' ? 'item' : 'criatura'
  return group(MOCK_DROPS.filter((d) => norm(d[by]).includes(norm(termo))), by, other)
}

/**
 * Sugestões para o autocomplete. Aceita um `signal` (AbortController) para
 * cancelar a requisição anterior quando o usuário continua digitando.
 */
export async function suggest(tipo, termo, signal) {
  if (!USE_MOCK) {
    const qs = new URLSearchParams({ tipo, q: termo })
    return callApi(`/api/sugestoes?${qs}`, { signal })
  }
  const by = campo(tipo)
  const nomes = [...new Set(MOCK_DROPS.map((d) => d[by]))]
  const t = norm(termo)
  return nomes
    .filter((n) => norm(n).includes(t))
    .sort((a, b) => Number(!norm(a).startsWith(t)) - Number(!norm(b).startsWith(t)))
    .slice(0, 12)
}

/** Envia o texto do Hunt Analyser e recebe os itens de task encontrados. */
export async function analyzeLog(log) {
  if (!USE_MOCK) {
    return callApi('/api/log', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ log }),
    })
  }
  await wait(300)
  if (!/loot/i.test(log)) throw new Error('Isso não parece um log do Hunt Analyser (sem seção de loot).')
  const taskItems = new Set(MOCK_DROPS.map((d) => norm(d.item)))
  const itens = []
  for (const m of log.matchAll(/^\s*(\d+)x\s+(?:an?\s+|the\s+)?(.+?)\s*$/gim)) {
    if (taskItems.has(norm(m[2]))) itens.push({ nome: m[2], quantidade: Number(m[1]) })
  }
  return itens
}
