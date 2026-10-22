import { slugItem } from './slug'

/**
 * Imagens dos itens: cada arquivo em src/assets/itens/<nome-do-item>.gif
 * (o nome vem de slugItem). O Vite lista os arquivos que existem na hora do
 * build, então um item sem imagem não gera nenhum pedido ao servidor: o site
 * só mostra um quadradinho no lugar.
 */
const arquivos = import.meta.glob('../assets/itens/*.gif', { eager: true, query: '?url', import: 'default' })

const urlPorSlug = new Map(
  Object.entries(arquivos).map(([caminho, url]) => [caminho.split('/').pop().replace(/\.gif$/, ''), url]),
)

/** Endereço da imagem do item, ou null se ainda não houver imagem. */
export function imagemItem(nome) {
  return urlPorSlug.get(slugItem(nome)) ?? null
}

/** Se existe ao menos uma imagem (o crédito à Tibia Wiki só aparece nesse caso). */
export const temImagensDeItens = urlPorSlug.size > 0
