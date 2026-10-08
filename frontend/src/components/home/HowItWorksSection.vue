<script setup>
const steps = [
  'Encerre sua caçada e abra o Hunt Analyser no Tibia.',
  'Copie o log e cole na página do Analisador.',
  'Veja os itens de task coletados e planeje a próxima hunt.',
]
</script>

<template>
  <section aria-labelledby="como-titulo" class="mx-auto max-w-3xl px-4 py-12">
    <h2 id="como-titulo" class="text-center text-3xl font-bold">Como funciona</h2>
    <!-- <ol> = lista ordenada: semântica correta para passos em sequência. -->
    <ol class="mt-8 space-y-4">
      <li v-for="(s, i) in steps" :key="s" class="card flex items-center gap-4">
        <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-gold font-bold text-bg">{{ i + 1 }}</span>
        <span>{{ s }}</span>
      </li>
    </ol>
  </section>
</template>
