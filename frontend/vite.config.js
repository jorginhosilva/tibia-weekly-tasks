import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  // Valor fixado na hora do build e usado no rodapé (veja AppFooter.vue).
  define: { __ANO_DO_BUILD__: JSON.stringify(new Date().getFullYear()) },
  resolve: {
    // Permite importar com "@/components/..." em vez de "../../components/..."
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  // Quando rodar `npm run dev`, as chamadas /api vão para o servidor Go.
  server: { proxy: { '/api': 'http://localhost:8080' } },
  // Opções do vite-ssg: gera um HTML pronto (pré-renderizado) para cada rota.
  // formatting 'none': o HTML pré-renderizado fica idêntico ao que o Vue gera no navegador.
  // ('minify' removia espaços entre textos e o Vue precisava corrigir isso ao hidratar.)
  ssgOptions: { script: 'async', formatting: 'none' },
})
