import { ViteSSG } from 'vite-ssg'
import App from './App.vue'
import { routes } from './router/routes'
import './style.css'

/**
 * Ponto de entrada. `ViteSSG` cria o app Vue + router e, no `npm run build`,
 * gera um arquivo HTML estático por rota (ótimo para SEO e velocidade).
 */
export const createApp = ViteSSG(
  App,
  {
    routes,
    // Ao trocar de página, volta ao topo (a menos que haja #âncora).
    scrollBehavior: (to) => (to.hash ? { el: to.hash, behavior: 'smooth' } : { top: 0 }),
  },
)
