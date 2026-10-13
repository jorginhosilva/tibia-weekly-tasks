<script setup>
import { useSeo } from '@/composables/useSeo'
import { SITE, PARTNER } from '@/config'

useSeo({
  title: 'Política de Privacidade',
  description: `Como o ${SITE.name} trata seus dados: não pedimos cadastro e não guardamos os logs que você cola.`,
  path: '/privacidade',
})
</script>

<template>
  <div class="mx-auto max-w-3xl px-4 py-12">
    <h1 class="text-3xl font-bold sm:text-4xl">Política de Privacidade</h1>
    <p class="mt-2 text-sm text-muted">Em conformidade com a Lei Geral de Proteção de Dados (LGPD).</p>

    <div class="mt-8 space-y-8 leading-relaxed text-muted">
      <section aria-labelledby="p-dados">
        <h2 id="p-dados" class="text-xl font-semibold text-text">O que coletamos</h2>
        <p class="mt-2">
          Não exigimos cadastro e não pedimos nome, e-mail ou qualquer dado pessoal. O texto do
          Hunt Analyser que você cola é enviado ao servidor <strong class="text-text">apenas para ser analisado
          naquele momento</strong> e não é armazenado.
        </p>
      </section>

      <section aria-labelledby="p-logs">
        <h2 id="p-logs" class="text-xl font-semibold text-text">Registros técnicos</h2>
        <p class="mt-2">
          Como todo site, o servidor e a hospedagem podem registrar dados técnicos de acesso (como endereço IP e
          horário) por motivos de segurança e para evitar abusos. Esses registros não são usados para identificar você.
        </p>
      </section>

      <section aria-labelledby="p-cookies">
        <h2 id="p-cookies" class="text-xl font-semibold text-text">Cookies</h2>
        <p class="mt-2">
          Este site não usa cookies de rastreamento nem de publicidade. O site do parceiro, ao qual você chega pelo
          botão de compra, tem as próprias regras (veja "Links de parceiros").
        </p>
      </section>

      <section aria-labelledby="p-links">
        <h2 id="p-links" class="text-xl font-semibold text-text">Links de parceiros</h2>
        <p class="mt-2">
          O botão "{{ PARTNER.label }}" leva à loja <strong class="text-text">{{ PARTNER.name }}</strong>, parceira do
          site, por meio de um link de afiliado. Se você usar o cupom
          <strong class="text-text">{{ PARTNER.cupom }}</strong> no pagamento, o site
          <strong class="text-text">pode receber uma comissão, sem custo adicional para você</strong>.
        </p>
        <p class="mt-2">
          Ao clicar no botão, o cupom é copiado para a área de transferência do seu aparelho (só isso; nada é enviado
          a ninguém) e você sai do nosso site. O {{ PARTNER.name }} tem a própria política de privacidade, da qual não
          somos responsáveis. Não enviamos a ele nenhum dado seu: o que você digita nas buscas e os logs que cola
          ficam apenas no nosso servidor.
        </p>
      </section>

      <section aria-labelledby="p-marca">
        <h2 id="p-marca" class="text-xl font-semibold text-text">Marcas</h2>
        <p class="mt-2">
          Tibia é marca registrada da CipSoft GmbH. Este é um projeto de fãs, sem vínculo com a empresa.
        </p>
      </section>
    </div>
  </div>
</template>
