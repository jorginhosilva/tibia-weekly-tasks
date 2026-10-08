import { ref } from 'vue'

/**
 * Envolve uma função assíncrona e expõe os 3 estados que toda tela precisa:
 * carregando, erro e dados. Evita repetir try/catch em cada página.
 *
 *   const { run, loading, error, data } = useAsyncAction(search)
 *   await run('criatura', 'bashmu')
 *
 * Se `run` for chamada de novo antes da anterior terminar, só a ÚLTIMA chamada
 * vale: uma resposta lenta e antiga nunca sobrescreve a resposta mais nova.
 */
export function useAsyncAction(fn) {
  const loading = ref(false)
  const error = ref('')
  const data = ref(null) // null = ainda não executou; [] = executou sem resultados

  let ultimaChamada = 0 // número da chamada mais recente

  async function run(...args) {
    const minhaChamada = ++ultimaChamada
    loading.value = true
    error.value = ''
    try {
      const resultado = await fn(...args)
      if (minhaChamada !== ultimaChamada) return // já existe uma chamada mais nova: descarta
      data.value = resultado
    } catch (e) {
      if (minhaChamada !== ultimaChamada) return
      data.value = null
      error.value = e.message
    } finally {
      if (minhaChamada === ultimaChamada) loading.value = false
    }
  }
  return { run, loading, error, data }
}
