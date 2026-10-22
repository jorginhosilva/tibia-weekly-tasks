<script setup>
import { SITE, PARTNER } from '@/config'
import { temImagensDeItens } from '@/utils/itens'
// Ano em que o site foi compilado (definido em vite.config.js). Usar o ano "de agora"
// faria o HTML pré-renderizado divergir do navegador quando o ano virasse.
const year = __ANO_DO_BUILD__
</script>

<template>
  <footer class="mt-20 border-t border-line bg-surface">
    <div class="mx-auto max-w-6xl px-4 py-8 text-sm text-muted">
      <p>© {{ year }} {{ SITE.name }}. Projeto de fãs, sem vínculo com a CipSoft GmbH.</p>
      <p class="mt-1">Tibia é marca registrada da CipSoft GmbH.</p>
      <!-- Só aparece quando o site tem imagens de itens (veja src/assets/itens). -->
      <p v-if="temImagensDeItens" class="mt-1">
        Imagens dos itens:
        <a href="https://www.tibiawiki.com.br" target="_blank" rel="noopener" class="underline hover:text-gold">Tibia Wiki</a>
        (arte do jogo, da CipSoft GmbH).
      </p>
      <p class="mt-1">
        O botão "{{ PARTNER.label }}" leva ao {{ PARTNER.name }}, parceiro do site. Se você usar o cupom
        <strong class="text-text">{{ PARTNER.cupom }}</strong> no pagamento, o site pode receber uma comissão,
        sem custo extra para você.
        <RouterLink to="/privacidade" class="underline hover:text-gold">Política de Privacidade</RouterLink>
      </p>
    </div>
  </footer>
</template>
