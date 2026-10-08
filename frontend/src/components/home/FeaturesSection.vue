<script setup>
// Dados separados do HTML: para editar o texto, não precisa mexer no template.
const features = [
  { icon: '🔎', title: 'Busca por criatura', text: 'Digite o nome de um monstro e veja todos os itens de task que ele dropa.' },
  { icon: '🎒', title: 'Busca por item', text: 'Sabe o item mas não onde caçar? Descubra quais criaturas o dropam.' },
  { icon: '📋', title: 'Analisador de log', text: 'Cole o log do Hunt Analyser e receba só os itens que importam para a task.' },
]
</script>

<template>
  <section aria-labelledby="recursos-titulo" class="mx-auto max-w-6xl px-4 py-12">
    <h2 id="recursos-titulo" class="text-center text-3xl font-bold">O que você encontra aqui</h2>
    <!-- Grid responsivo: 1 coluna no celular, 3 a partir de telas médias. -->
    <ul class="mt-10 grid gap-5 md:grid-cols-3">
      <li v-for="f in features" :key="f.title" class="card transition-transform hover:-translate-y-1">
        <span class="text-3xl" aria-hidden="true">{{ f.icon }}</span>
        <h3 class="mt-3 text-lg font-semibold">{{ f.title }}</h3>
        <p class="mt-2 text-muted">{{ f.text }}</p>
      </li>
    </ul>
  </section>
</template>
