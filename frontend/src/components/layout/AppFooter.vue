<script setup>
import { SITE, PARTNER } from '@/config'
// Ano em que o site foi compilado (definido em vite.config.js). Usar o ano "de agora"
// faria o HTML pré-renderizado divergir do navegador quando o ano virasse.
const year = __ANO_DO_BUILD__
</script>

<template>
  <footer class="mt-20 border-t border-line bg-surface">
    <div class="mx-auto max-w-6xl px-4 py-8 text-sm text-muted">
      <p>© {{ year }} {{ SITE.name }}. Projeto de fãs, sem vínculo com a CipSoft GmbH.</p>
      <p class="mt-1">Tibia é marca registrada da CipSoft GmbH.</p>
      <p class="mt-1">
        O botão "{{ PARTNER.label }}" é um link de afiliado do {{ PARTNER.name }}: se você comprar por ele,
        o site pode receber uma comissão, sem custo extra para você.
        <RouterLink to="/privacidade" class="underline hover:text-gold">Política de Privacidade</RouterLink>
      </p>
    </div>
  </footer>
</template>
