<script setup>
/**
 * Cartão de um NPC: o nome, quantos itens ele compra e a grade "item + preço".
 * A comparação com o Market fica com quem usa o site: aqui só mostramos o preço do NPC.
 */
defineProps({
  nome: { type: String, required: true },
  precos: { type: Array, required: true }, // [{ nome: 'Chaos Mace', preco: 9000 }]
})

const gold = new Intl.NumberFormat('pt-BR') // 9000 -> "9.000"
</script>

<template>
  <article class="card">
    <header class="flex items-center justify-between gap-3">
      <!-- h2: logo abaixo do h1 da página, sem pular nível (importante para leitores de tela). -->
      <h2 class="text-lg font-semibold text-gold">{{ nome }}</h2>
      <span class="rounded-full border border-line px-3 py-1 text-xs text-muted">{{ precos.length }} itens</span>
    </header>
    <p class="mt-1 text-xs uppercase tracking-wide text-muted">Quanto paga por item</p>
    <ul class="mt-3 grid gap-px overflow-hidden rounded-lg border border-line bg-line sm:grid-cols-2">
      <li v-for="p in precos" :key="p.nome" class="flex items-baseline justify-between gap-3 bg-surface-2 px-3 py-2 text-sm">
        <span>{{ p.nome }}</span>
        <span class="shrink-0 font-semibold text-gold">{{ gold.format(p.preco) }} gp</span>
      </li>
    </ul>
  </article>
</template>
