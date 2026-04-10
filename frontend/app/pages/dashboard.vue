<script setup lang="ts">
definePageMeta({
  middleware: "auth",
});

await navigateTo("/", { redirectCode: 301, replace: true });
</script>
