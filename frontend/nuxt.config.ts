export default defineNuxtConfig({
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  srcDir: "app/",
  modules: ["@pinia/nuxt", "@vueuse/nuxt"],
  css: ["~/assets/css/main.css"],
  runtimeConfig: {
    apiBaseUrl:
      process.env.NUXT_API_BASE_URL ||
      process.env.NUXT_PUBLIC_API_BASE_URL ||
      "http://localhost:8080/api/v1",
    public: {
      apiBaseUrl:
        process.env.NUXT_PUBLIC_API_BASE_URL || "http://localhost:8080/api/v1",
      appName: process.env.NUXT_PUBLIC_APP_NAME || "IMS Thailand",
    },
  },
  typescript: {
    strict: true,
  },
  imports: {
    dirs: ["composables", "stores"],
  },
  components: [
    {
      path: "~/shared/ui",
      pathPrefix: false,
    },
  ],
  postcss: {
    plugins: {
      "@tailwindcss/postcss": {},
      autoprefixer: {},
    },
  },
  vite: {
    server: {
      fs: {
        strict: false,
      },
      watch: {
        usePolling: true,
        interval: 1000,
      },
    },
  },
});
