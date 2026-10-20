<script setup>
/** Cartão de um resultado: um nome (criatura/item) e a lista de relacionados. */
defineProps({
  nome: { type: String, required: true },
  relacionados: { type: Array, required: true },
  rotulo: { type: String, required: true }, // ex.: "Itens de task"
  // Só na busca por item: os NPCs que compram o item e por quanto.
  precos: { type: Array, default: () => [] }, // [{ nome: 'Rashid', preco: 9000 }]
})

const gold = new Intl.NumberFormat('pt-BR') // 9000 -> "9.000"
</script>

<template>
  <article class="card">
    <!-- h2: logo abaixo do h1 da página, sem pular nível (importante para leitores de tela). -->
    <h2 class="text-lg font-semibold text-gold">{{ nome }}</h2>
    <p class="mt-1 text-xs uppercase tracking-wide text-muted">{{ rotulo }}</p>
    <ul class="mt-3 flex flex-wrap gap-2">
      <li v-for="r in relacionados" :key="r" class="rounded-full border border-line bg-surface-2 px-3 py-1 text-sm">{{ r }}</li>
    </ul>
    <template v-if="precos.length">
      <p class="mt-4 text-xs uppercase tracking-wide text-muted">NPCs que compram</p>
      <ul class="mt-2 flex flex-wrap gap-2">
        <li v-for="p in precos" :key="p.nome" class="rounded-full border border-gold/40 bg-surface-2 px-3 py-1 text-sm">
          {{ p.nome }} · <span class="font-semibold text-gold">{{ gold.format(p.preco) }} gp</span>
        </li>
      </ul>
    </template>
  </article>
</template>
