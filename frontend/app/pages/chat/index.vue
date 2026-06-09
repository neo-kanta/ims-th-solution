<script setup lang="ts">
import ChatWindow from "~/features/chat/components/ChatWindow.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth"],
});
</script>

<template>
  <ChatWindow />
</template>
