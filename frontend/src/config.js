/**
 * Configurações globais do site. Mude aqui e todas as páginas (SEO incluso) acompanham.
 */
export const SITE = {
  name: 'Tibia Weekly Tasks',
  url: 'https://tibiatask.com.br', // sem barra no final
  description:
    'Descubra quais criaturas dropam os itens da Weekly Task do Tibia e analise o log do Hunt Analyser para saber o que já coletou.',
  locale: 'pt_BR',
  // Imagem que aparece ao compartilhar o link (WhatsApp, Discord...). 1200x630.
  // O arquivo fica em public/og-image.png; o desenho editável, em design/og-image.svg.
  ogImage: {
    path: '/og-image.png',
    width: 1200,
    height: 630,
    alt: 'Tibia Weekly Tasks: descubra qual criatura dropa cada item da task e analise o log do Hunt Analyser.',
  },
}

/**
 * Botão de parceria exibido no cabeçalho: link do Mikicoins.
 * A comissão só conta se o cliente usar o `cupom` no pagamento. Por isso, clicar no
 * botão copia o cupom e avisa para colá-lo no checkout (veja AppHeader.vue).
 * Para trocar de parceiro, mude os campos abaixo.
 */
export const PARTNER = {
  name: 'Mikicoins',
  label: 'Comprar Tibia Coins',
  url: 'https://mikicoins.com/',
  cupom: 'TIBIATASK', // código de afiliado que o cliente cola no pagamento
}
