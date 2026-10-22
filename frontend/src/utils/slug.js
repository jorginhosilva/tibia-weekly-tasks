/**
 * Nome do arquivo da imagem de um item (sem a extensão .gif):
 * "Spellweaver's Robe" -> "spellweaver-s-robe", "Darklight Core (Item)" -> "darklight-core-item".
 *
 * Fica num arquivo próprio, sem nada do Vite, para poder ser usado também
 * fora do site (por exemplo, num script em Node que organize novas imagens).
 */
export function slugItem(nome) {
  return nome
    .normalize('NFD')
    .replace(/\p{M}/gu, '') // tira acentos
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
}
