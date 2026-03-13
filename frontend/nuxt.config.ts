// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  srcDir: 'app/',
  modules: [
    '@pinia/nuxt',
    '@vueuse/nuxt',
  ],
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    public: {
      apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1',
      appName: process.env.NUXT_PUBLIC_APP_NAME || 'IMS Thailand',
    },
  },
  typescript: {
    strict: true,
  },
  imports: {
    dirs: [
      '../shared/stores',
      '../shared/composables',
      '../shared/utils',
    ],
  },
  postcss: {
    plugins: {
      '@tailwindcss/postcss': {},
      autoprefixer: {},
    },
  },
})
