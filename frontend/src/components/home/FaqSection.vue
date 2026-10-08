<script setup>
import { useHead } from '@unhead/vue'

// Perguntas frequentes: aparecem na página E viram dados estruturados (FAQPage),
// o que pode gerar resultados expandidos no Google.
const faqs = [
  { q: 'O que é a Weekly Task do Tibia?', a: 'É uma missão semanal em que você caça criaturas específicas e coleta itens para receber recompensas.' },
  { q: 'Como uso o Hunt Analyser?', a: 'Ao fim da caçada, copie o log do Hunt Analyser e cole na página Analisador de Log.' },
  { q: 'Preciso criar conta?', a: 'Não. Tudo funciona sem cadastro e nenhum dado seu é armazenado.' },
]

useHead({
  script: [{
    type: 'application/ld+json',
    innerHTML: JSON.stringify({
      '@context': 'https://schema.org',
      '@type': 'FAQPage',
      mainEntity: faqs.map((f) => ({
        '@type': 'Question', name: f.q,
        acceptedAnswer: { '@type': 'Answer', text: f.a },
      })),
    }),
  }],
})
</script>

<template>
  <section id="faq" aria-labelledby="faq-titulo" class="mx-auto max-w-3xl scroll-mt-20 px-4 py-12">
    <h2 id="faq-titulo" class="text-center text-3xl font-bold">Perguntas frequentes</h2>
    <div class="mt-8 space-y-3">
      <!-- <details>/<summary> é um acordeão nativo: acessível e sem JavaScript. -->
      <details v-for="f in faqs" :key="f.q" class="card">
        <summary class="cursor-pointer font-semibold">{{ f.q }}</summary>
        <p class="mt-3 text-muted">{{ f.a }}</p>
      </details>
    </div>
  </section>
</template>
