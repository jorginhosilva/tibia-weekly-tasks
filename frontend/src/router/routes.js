/**
 * Mapa de rotas. Cada página é carregada sob demanda (`() => import(...)`),
 * então o visitante só baixa o código da página que abrir (code splitting).
 */
export const routes = [
  { path: '/', name: 'home', component: () => import('@/pages/HomePage.vue') },
  { path: '/buscar', name: 'buscar', component: () => import('@/pages/SearchPage.vue') },
  { path: '/analisador-de-log', name: 'log', component: () => import('@/pages/LogAnalyzerPage.vue') },
  { path: '/privacidade', name: 'privacidade', component: () => import('@/pages/PrivacyPage.vue') },
  // Rota fixa só para o build gerar o arquivo 404.html. O servidor Go o entrega,
  // com status 404, para qualquer endereço que não exista.
  { path: '/404', name: 'nao-encontrada', component: () => import('@/pages/NotFoundPage.vue') },
  { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('@/pages/NotFoundPage.vue') },
]
