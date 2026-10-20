<script setup>
import { ref, watch, computed, onBeforeUnmount } from 'vue'
import { suggest } from '@/services/tasksApi'

/**
 * Campo de busca com autocomplete.
 *
 * Segue o padrão ARIA "combobox + listbox": leitores de tela anunciam as
 * sugestões e o teclado funciona por completo:
 *   ↓ / ↑  navega pela lista     Enter  escolhe a opção destacada
 *   Esc    fecha a lista         (sem opção destacada, Enter faz a busca normal)
 *
 * Props:  v-model (texto digitado) e `tipo` ('criatura' | 'item' | 'npc')
 * Evento: `select` — disparado quando o usuário ESCOLHE uma sugestão
 */
const props = defineProps({
  modelValue: { type: String, default: '' },
  tipo: { type: String, required: true },
  placeholder: { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue', 'select'])

const MIN_CHARS = 2 // só sugere a partir de 2 letras (evita lista enorme e requisições inúteis)
const DEBOUNCE_MS = 200 // espera o usuário "pausar" a digitação antes de pedir sugestões

const sugestoes = ref([])
const aberta = ref(false)
const ativa = ref(-1) // índice da opção destacada (-1 = nenhuma)
const listId = 'sugestoes-lista'

let timer = null
let controller = null // permite cancelar a requisição anterior

const texto = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

/** Busca sugestões, ignorando respostas antigas (se o usuário já digitou mais). */
async function carregar() {
  controller?.abort()
  const termo = texto.value.trim()
  if (termo.length < MIN_CHARS) {
    sugestoes.value = []
    aberta.value = false
    return
  }
  controller = new AbortController()
  try {
    sugestoes.value = await suggest(props.tipo, termo, controller.signal)
    ativa.value = -1
    aberta.value = sugestoes.value.length > 0
  } catch (e) {
    if (e.name !== 'AbortError') aberta.value = false // sugestão é opcional: falha em silêncio
  }
}

function agendar() {
  clearTimeout(timer)
  timer = setTimeout(carregar, DEBOUNCE_MS)
}

// Nome da última sugestão escolhida (evita reabrir a lista logo após escolher).
let escolhido = ''

// Digitou? Reagenda a busca de sugestões.
watch(texto, () => {
  // O texto mudou por causa de uma escolha na lista: não reabrir a lista.
  if (escolhido !== '' && texto.value === escolhido) return
  // Qualquer outra mudança (apagar, digitar de novo) volta ao comportamento normal.
  escolhido = ''
  agendar()
})

// Trocou o tipo de busca (criatura/item/NPC)? As sugestões antigas não valem mais.
watch(
  () => props.tipo,
  () => {
    escolhido = ''
    agendar()
  },
)

function escolher(nome) {
  escolhido = nome
  texto.value = nome
  fechar()
  emit('select', nome)
}

function fechar() {
  clearTimeout(timer)
  controller?.abort()
  aberta.value = false
  ativa.value = -1
}

function onKeydown(e) {
  if (!aberta.value) return
  const n = sugestoes.value.length
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    ativa.value = (ativa.value + 1) % n
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    ativa.value = (ativa.value - 1 + n) % n
  } else if (e.key === 'Enter' && ativa.value >= 0) {
    e.preventDefault() // não envia o formulário: só escolhe a sugestão
    escolher(sugestoes.value[ativa.value])
  } else if (e.key === 'Enter') {
    fechar() // Enter sem opção destacada: o formulário faz a busca normal
  } else if (e.key === 'Escape') {
    fechar()
  }
}

/** Divide o nome em [antes, trecho digitado, depois] para destacar o trecho em negrito. */
function partes(nome) {
  const termo = texto.value.trim()
  const i = nome.toLowerCase().indexOf(termo.toLowerCase())
  if (i < 0 || !termo) return [nome, '', '']
  return [nome.slice(0, i), nome.slice(i, i + termo.length), nome.slice(i + termo.length)]
}

onBeforeUnmount(fechar)
</script>

<template>
  <div class="relative w-full">
    <label for="busca" class="sr-only">Nome</label>
    <input
      id="busca"
      v-model="texto"
      class="input"
      type="text"
      maxlength="100"
      autocomplete="off"
      role="combobox"
      aria-autocomplete="list"
      :aria-expanded="aberta"
      :aria-controls="listId"
      :aria-activedescendant="ativa >= 0 ? `${listId}-${ativa}` : undefined"
      :placeholder="placeholder"
      @keydown="onKeydown"
      @blur="fechar"
    />

    <!-- mousedown.prevent: impede o input de perder o foco (e fechar a lista) antes do clique valer. -->
    <ul
      :id="listId"
      role="listbox"
      aria-label="Sugestões"
      :class="{ hidden: !aberta }"
      class="absolute left-0 right-0 top-full z-30 mt-1 max-h-72 overflow-auto rounded-lg border border-line bg-surface shadow-xl"
    >
      <li
        v-for="(nome, i) in sugestoes"
        :id="`${listId}-${i}`"
        :key="nome"
        role="option"
        :aria-selected="i === ativa"
        class="cursor-pointer px-4 py-3 text-sm"
        :class="i === ativa ? 'bg-surface-2 text-gold' : 'hover:bg-surface-2'"
        @mousedown.prevent="escolher(nome)"
        @mousemove="ativa = i"
      >
        {{ partes(nome)[0] }}<strong class="text-gold">{{ partes(nome)[1] }}</strong>{{ partes(nome)[2] }}
      </li>
    </ul>

    <!-- Anuncia para leitores de tela quantas sugestões existem. -->
    <p class="sr-only" aria-live="polite">
      {{ aberta ? `${sugestoes.length} sugestões disponíveis` : '' }}
    </p>
  </div>
</template>
