<script setup>
import AppHeader from '@/components/layout/AppHeader.vue'
import AppFooter from '@/components/layout/AppFooter.vue'
</script>

<template>
  <!-- Link "pular para o conteúdo": ajuda leitores de tela e navegação por teclado. -->
  <a
    href="#conteudo"
    class="sr-only z-50 rounded bg-gold px-4 py-2 font-semibold text-bg focus:not-sr-only focus:fixed focus:left-4 focus:top-4"
  >
    Pular para o conteúdo
  </a>
  <AppHeader />
  <main id="conteudo" tabindex="-1" class="min-h-[70vh]">
    <RouterView />
  </main>
  <AppFooter />
</template>
