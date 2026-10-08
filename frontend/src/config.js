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
 * Botão de parceria exibido no cabeçalho: link de afiliado do Moedaz.
 * Para trocar de parceiro, mude a `url` (e o `label`, se quiser).
 */
export const PARTNER = {
  name: 'Moedaz',
  label: 'Comprar Tibia Coins',
  url: 'https://www.moedaz.com/?ref=2RAH2ZQDT227KTPO',
}
