<script setup>
import { ref } from 'vue'
import StatusMessage from '@/components/search/StatusMessage.vue'
import { useSeo } from '@/composables/useSeo'
import { useAsyncAction } from '@/composables/useAsyncAction'
import { analyzeLog } from '@/services/tasksApi'
import { MOCK_LOG } from '@/data/mockData'

useSeo({
  title: 'Analisador de log do Hunt Analyser',
  description: 'Cole o log do Hunt Analyser do Tibia e veja quais itens da Weekly Task você já coletou.',
  path: '/analisador-de-log',
})

const MAX = 64 * 1024 // mesmo limite do servidor Go (64 KB)
const log = ref('')
const { run, loading, error, data } = useAsyncAction(analyzeLog)

async function onSubmit() {
  if (!log.value.trim()) {
    data.value = null
    error.value = 'Cole o log antes de analisar.'
    return
  }
  if (log.value.length > MAX) {
    data.value = null
    error.value = 'Texto grande demais (máximo 64 KB).'
    return
  }
  await run(log.value)
}
</script>

<template>
  <div class="mx-auto max-w-3xl px-4 py-12">
    <h1 class="text-3xl font-bold sm:text-4xl">Analisador de log</h1>
    <p class="mt-2 text-muted">Cole o log do Hunt Analyser e veja só os itens de task.</p>

    <form class="mt-8 space-y-4" @submit.prevent="onSubmit">
      <label for="log" class="block font-semibold">Log do Hunt Analyser</label>
      <textarea id="log" v-model="log" rows="10" class="input py-3 font-mono text-sm" placeholder="Cole aqui o log..."></textarea>
      <div class="flex flex-wrap gap-3">
        <button class="btn btn-primary" type="submit" :disabled="loading">{{ loading ? 'Analisando…' : 'Analisar log' }}</button>
        <button class="btn btn-ghost" type="button" @click="log = MOCK_LOG">Usar exemplo</button>
      </div>
    </form>

    <div class="mt-8" aria-live="polite">
      <StatusMessage v-if="error" type="error">{{ error }}</StatusMessage>
      <StatusMessage v-else-if="data && !data.length">Nenhum item de weekly task neste log.</StatusMessage>
      <section v-else-if="data" class="card" aria-labelledby="res-titulo">
        <h2 id="res-titulo" class="text-xl font-semibold text-gold">Itens de task encontrados</h2>
        <ul class="mt-4 divide-y divide-line">
          <li v-for="i in data" :key="i.nome" class="flex justify-between py-2">
            <span>{{ i.nome }}</span><span class="font-bold text-gold">{{ i.quantidade }}x</span>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>
