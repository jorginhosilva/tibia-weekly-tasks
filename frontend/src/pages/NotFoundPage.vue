<script setup>
import { useHead } from '@unhead/vue'
import { useSeo } from '@/composables/useSeo'

useSeo({ title: 'Página não encontrada', path: '/404' })
// noindex: pede ao Google para não indexar a página de erro.
useHead({ meta: [{ name: 'robots', content: 'noindex' }] })
</script>

<template>
  <div class="mx-auto max-w-xl px-4 py-24 text-center">
    <h1 class="text-5xl font-extrabold text-gold">404</h1>
    <p class="mt-4 text-muted">Essa página não existe.</p>
    <RouterLink to="/" class="btn btn-primary mt-8">Voltar ao início</RouterLink>
  </div>
</template>
