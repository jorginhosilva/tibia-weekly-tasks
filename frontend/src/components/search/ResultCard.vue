<script setup>
/** Cartão de um resultado: um nome (criatura/item) e a lista de relacionados. */
defineProps({
  nome: { type: String, required: true },
  relacionados: { type: Array, required: true },
  rotulo: { type: String, required: true }, // ex.: "Itens de task"
})
</script>

<template>
  <article class="card">
    <!-- h2: logo abaixo do h1 da página, sem pular nível (importante para leitores de tela). -->
    <h2 class="text-lg font-semibold text-gold">{{ nome }}</h2>
    <p class="mt-1 text-xs uppercase tracking-wide text-muted">{{ rotulo }}</p>
    <ul class="mt-3 flex flex-wrap gap-2">
      <li v-for="r in relacionados" :key="r" class="rounded-full border border-line bg-surface-2 px-3 py-1 text-sm">{{ r }}</li>
    </ul>
  </article>
</template>
