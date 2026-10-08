<template>
  <section aria-labelledby="hero-titulo" class="relative overflow-hidden">
    <!-- Brilho decorativo (aria-hidden: leitores de tela ignoram). -->
    <div aria-hidden="true" class="pointer-events-none absolute -top-40 left-1/2 h-96 w-[40rem] -translate-x-1/2 rounded-full bg-gold/10 blur-3xl"></div>
    <div class="relative mx-auto max-w-4xl px-4 py-16 text-center sm:py-24">
      <p class="mb-4 inline-block rounded-full border border-line bg-surface px-3 py-1 text-xs text-gold">
        ⚔️ Ferramenta gratuita para jogadores de Tibia
      </p>
      <!-- Só UM <h1> por página: é o título principal para o Google. -->
      <h1 id="hero-titulo" class="text-4xl font-extrabold leading-tight tracking-tight sm:text-6xl">
        Complete sua <span class="text-gold">Weekly Task</span> mais rápido
      </h1>
      <p class="mx-auto mt-5 max-w-2xl text-lg text-muted">
        Descubra qual criatura dropa cada item da task e cole o log do Hunt Analyser para
        ver na hora quanto você já coletou.
      </p>
      <div class="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
        <RouterLink to="/buscar" class="btn btn-primary">Buscar criatura ou item</RouterLink>
        <RouterLink to="/analisador-de-log" class="btn btn-ghost">Analisar meu log</RouterLink>
      </div>
    </div>
  </section>
</template>
