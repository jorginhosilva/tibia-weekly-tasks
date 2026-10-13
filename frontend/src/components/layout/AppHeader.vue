<script setup>
import { ref } from 'vue'
import { SITE, PARTNER } from '@/config'

// Links do menu. Adicionar uma página = adicionar uma linha aqui.
const links = [
  { to: '/buscar', label: 'Buscar' },
  { to: '/analisador-de-log', label: 'Analisador de Log' },
  { to: '/#faq', label: 'Dúvidas' },
]
const open = ref(false) // menu mobile aberto?

// Aviso "cupom copiado", mostrado depois do clique no botão de parceria.
const copiado = ref(false)
let timer

/**
 * Copia o cupom ao clicar no botão de parceria. O link abre normalmente (não usamos
 * preventDefault); a cópia só ajuda o cliente a colar o cupom no pagamento.
 * Se o navegador negar a cópia, o aviso aparece do mesmo jeito, com o cupom escrito.
 */
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
</script>

<template>
  <header class="sticky top-0 z-40 border-b border-line bg-bg/90 backdrop-blur">
    <div class="mx-auto flex h-16 max-w-6xl items-center justify-between px-4">
      <RouterLink to="/" class="flex items-center gap-2 text-lg font-bold" @click="open = false">
        <img src="/favicon.svg" alt="" width="32" height="32" />
        <!-- No celular só o ícone aparece, para sobrar espaço ao botão de parceria. -->
        <span class="sr-only sm:not-sr-only">{{ SITE.name }}</span>
      </RouterLink>

      <!-- Botão de parceria. rel="sponsored" avisa o Google que é link comercial/afiliado.
           Sem "noreferrer" de propósito: alguns programas de afiliados conferem de qual site veio a visita. -->
      <a
        :href="PARTNER.url" target="_blank" rel="sponsored noopener"
        :title="`Link de afiliado: ${PARTNER.name}`"
        class="btn btn-primary rounded-full px-6 shadow-lg shadow-gold/20 sm:min-w-72 sm:px-10"
        @click="copiarCupom"
      >
        <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
          <circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" stroke-width="2" />
          <path d="M12 7v10M9.5 9.5h4a1.5 1.5 0 0 1 0 3h-3a1.5 1.5 0 0 0 0 3h4" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
        </svg>
        {{ PARTNER.label }}
        <span class="sr-only">(link de afiliado do {{ PARTNER.name }}, abre em nova aba; copia o cupom {{ PARTNER.cupom }})</span>
      </a>

      <!-- Aviso do cupom: fica solto embaixo do cabeçalho, para a barra não mudar de altura.
           role="status": leitores de tela anunciam quando aparece. -->
      <div class="pointer-events-none absolute inset-x-0 top-full flex justify-center px-4">
        <p
          v-if="copiado" role="status"
          class="pointer-events-auto mt-2 max-w-md rounded-xl border border-gold/50 bg-surface px-4 py-3 text-center text-sm shadow-xl shadow-black/40"
        >
          Cupom <strong class="rounded bg-surface-2 px-2 py-0.5 text-gold">{{ PARTNER.cupom }}</strong> copiado!
          <span class="block text-muted">Cole no pagamento no {{ PARTNER.name }} para apoiar o site.</span>
        </p>
      </div>

      <!-- Menu desktop -->
      <nav aria-label="Principal" class="hidden gap-6 md:flex">
        <RouterLink
          v-for="l in links" :key="l.to" :to="l.to"
          class="text-sm text-muted transition-colors hover:text-gold"
          :active-class="l.to.includes('#') ? '' : '!text-gold'"
        >{{ l.label }}</RouterLink>
      </nav>

      <!-- Botão hambúrguer (só no mobile). aria-expanded informa o estado a leitores de tela. -->
      <button
        class="btn btn-ghost md:hidden" type="button"
        :aria-expanded="open" aria-controls="menu-mobile"
        @click="open = !open"
      >
        <span class="sr-only">Abrir menu</span>
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
          <path :d="open ? 'M6 6l12 12M18 6 6 18' : 'M4 6h16M4 12h16M4 18h16'" stroke-linecap="round" />
        </svg>
      </button>
    </div>

    <nav v-if="open" id="menu-mobile" aria-label="Principal (mobile)" class="border-t border-line md:hidden">
      <RouterLink
        v-for="l in links" :key="l.to" :to="l.to"
        class="block px-4 py-3 text-muted hover:bg-surface hover:text-gold"
        @click="open = false"
      >{{ l.label }}</RouterLink>
    </nav>
  </header>
</template>
