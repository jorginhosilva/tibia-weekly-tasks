<script setup>
/** Mensagem de erro/vazio. role="alert" faz leitores de tela anunciarem o erro. */
defineProps({ type: { type: String, default: 'info' } }) // 'info' | 'error'
</script>

<template>
  <p
    :role="type === 'error' ? 'alert' : 'status'"
    class="card text-center"
    :class="type === 'error' ? 'border-danger/50 text-danger' : 'text-muted'"
  >
    <slot />
  </p>
</template>
