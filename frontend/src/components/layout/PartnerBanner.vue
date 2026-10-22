<script setup>
/**
 * Imagem do parceiro (Mikicoins). Ao clicar, abre o site dele e copia o cupom,
 * igual ao botão do cabeçalho: a comissão só conta se o cliente usar o cupom.
 * A imagem tem 500 px de largura, então nunca é esticada além disso (ficaria borrada).
 */
import { PARTNER } from '@/config'
import { useCupomParceiro } from '@/composables/useCupomParceiro'
import imagem from '@/assets/parceiro/mikicoins.jpg'

const { copiarCupom } = useCupomParceiro()
</script>

<template>
  <aside class="mx-auto w-full max-w-[500px]" aria-label="Parceiro">
    <!-- rel="sponsored": avisa o Google que é link comercial/afiliado (igual ao botão do topo). -->
    <a
      :href="PARTNER.url" target="_blank" rel="sponsored noopener"
      :title="`Link de afiliado: ${PARTNER.name}`"
      class="block overflow-hidden rounded-xl border border-line transition-colors hover:border-gold"
      @click="copiarCupom"
    >
      <img
        :src="imagem"
        width="500" height="182" loading="lazy"
        :alt="`${PARTNER.name}: compra e venda de Tibia Coins com entrega imediata`"
        class="block h-auto w-full"
      />
      <span class="sr-only">(link de afiliado, abre em nova aba; copia o cupom {{ PARTNER.cupom }})</span>
    </a>
    <p class="mt-2 text-center text-sm text-muted">
      Use o cupom <strong class="rounded bg-surface-2 px-2 py-0.5 text-gold">{{ PARTNER.cupom }}</strong> no pagamento
    </p>
  </aside>
</template>
