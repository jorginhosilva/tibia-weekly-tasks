<script setup>
import { ref, computed } from 'vue'
import AutocompleteInput from '@/components/search/AutocompleteInput.vue'
import ResultCard from '@/components/search/ResultCard.vue'
import StatusMessage from '@/components/search/StatusMessage.vue'
import { useSeo } from '@/composables/useSeo'
import { useAsyncAction } from '@/composables/useAsyncAction'
import { search } from '@/services/tasksApi'

useSeo({
  title: 'Buscar criatura ou item da Weekly Task',
  description: 'Pesquise qual criatura dropa cada item da Weekly Task do Tibia, ou quais itens uma criatura dropa.',
  path: '/buscar',
})

const tipos = [
  { v: 'criatura', t: 'Por criatura' },
  { v: 'item', t: 'Por item' },
]
const tipo = ref('criatura') // 'criatura' | 'item'
const termo = ref('')
const { run, loading, error, data } = useAsyncAction(search)
const buscado = ref({ tipo: '', termo: '' }) // o que foi realmente buscado

const rotulo = computed(() => (buscado.value.tipo === 'criatura' ? 'Itens de task' : 'Criaturas que dropam'))

async function onSubmit() {
  const t = termo.value.trim()
  if (!t) {
    data.value = null // some com os resultados antigos: só a mensagem de erro aparece
    error.value = 'Digite algo para buscar.'
    return
  }
  buscado.value = { tipo: tipo.value, termo: t }
  await run(tipo.value, t)
}
</script>

<template>
  <div class="mx-auto max-w-3xl px-4 py-12">
    <h1 class="text-3xl font-bold sm:text-4xl">Buscar na Weekly Task</h1>
    <p class="mt-2 text-muted">Escolha o tipo de busca e digite o nome.</p>

    <form class="mt-8 space-y-4" role="search" @submit.prevent="onSubmit">
      <fieldset class="flex gap-4">
        <legend class="sr-only">Tipo de busca</legend>
        <label v-for="o in tipos" :key="o.v" class="flex items-center gap-2">
          <input v-model="tipo" type="radio" name="tipo" :value="o.v" class="accent-gold" /> {{ o.t }}
        </label>
      </fieldset>
      <div class="flex flex-col gap-3 sm:flex-row">
        <!-- Ao escolher uma sugestão, já faz a busca (um clique a menos para o usuário). -->
        <AutocompleteInput
          v-model="termo"
          :tipo="tipo"
          :placeholder="tipo === 'criatura' ? 'ex.: deepworm' : 'ex.: bashmu fang'"
          @select="onSubmit"
        />
        <button class="btn btn-primary" type="submit" :disabled="loading">{{ loading ? 'Buscando…' : 'Buscar' }}</button>
      </div>
    </form>

    <!-- aria-live: leitores de tela anunciam quando os resultados mudam. -->
    <div class="mt-8 space-y-4" aria-live="polite">
      <StatusMessage v-if="error" type="error">{{ error }}</StatusMessage>
      <StatusMessage v-else-if="data && !data.length">Nada encontrado para “{{ buscado.termo }}”.</StatusMessage>
      <ResultCard v-for="r in data ?? []" :key="r.nome" :nome="r.nome" :relacionados="r.relacionados" :rotulo="rotulo" />
    </div>
  </div>
</template>
