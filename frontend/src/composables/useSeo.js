import { useHead, useSeoMeta } from '@unhead/vue'
import { SITE } from '@/config'

/**
 * Define as meta tags de SEO de uma página.
 *
 * Uso:  useSeo({ title: 'Buscar', description: '...', path: '/buscar' })
 *
 * - title/description: o que aparece no Google.
 * - path: caminho da página, usado para a URL canônica (evita conteúdo duplicado).
 * - schema: (opcional) objeto JSON-LD (dados estruturados) específico da página.
 */
export function useSeo({ title, description = SITE.description, path = '/', schema }) {
  const fullTitle = title ? `${title} | ${SITE.name}` : `${SITE.name} – Itens e criaturas da Weekly Task`
  const url = SITE.url + path
  const imagem = SITE.url + SITE.ogImage.path

  useHead({
    title: fullTitle,
    htmlAttrs: { lang: 'pt-BR' }, // o unhead sobrescreve o lang do index.html, então definimos aqui
    link: [{ rel: 'canonical', href: url }],
    script: schema
      ? [{ type: 'application/ld+json', innerHTML: JSON.stringify(schema) }]
      : [],
  })

  useSeoMeta({
    description,
    ogTitle: fullTitle,
    ogDescription: description,
    ogUrl: url,
    ogType: 'website',
    ogLocale: SITE.locale,
    ogSiteName: SITE.name,
    // A imagem precisa de endereço COMPLETO (com https://), senão as redes não a encontram.
    ogImage: imagem,
    ogImageType: 'image/png',
    ogImageWidth: SITE.ogImage.width,
    ogImageHeight: SITE.ogImage.height,
    ogImageAlt: SITE.ogImage.alt,
    twitterCard: 'summary_large_image',
    twitterTitle: fullTitle,
    twitterDescription: description,
    twitterImage: imagem,
    twitterImageAlt: SITE.ogImage.alt,
  })
}
