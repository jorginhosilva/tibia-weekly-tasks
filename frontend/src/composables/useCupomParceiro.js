import { ref } from 'vue'
import { PARTNER } from '@/config'

// Estado único para o site todo: o aviso "cupom copiado" aparece embaixo do
// cabeçalho, seja o clique no botão do topo ou na imagem do parceiro.
const copiado = ref(false)
let timer

/**
 * Copia o cupom do parceiro ao clicar num link dele. O link abre normalmente (não usamos
 * preventDefault); a cópia só ajuda o cliente a colar o cupom no pagamento.
 * Se o navegador negar a cópia, o aviso aparece do mesmo jeito, com o cupom escrito.
 */
export function useCupomParceiro() {
  async function copiarCupom() {
    try {
      await navigator.clipboard.writeText(PARTNER.cupom)
    } catch {
      // sem permissão para copiar: o cliente vê o cupom no aviso e digita
    }
    copiado.value = true
    clearTimeout(timer)
    timer = setTimeout(() => (copiado.value = false), 8000)
  }
  return { copiado, copiarCupom }
}
