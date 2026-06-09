<script setup lang="ts">
import type { ChatMessage } from "../composables/useChatStream";
import MessageBubble from "./MessageBubble.vue";

defineProps<{
  messages: readonly ChatMessage[];
  streaming: boolean;
}>();
</script>

<template>
  <div class="chat-list">
    <MessageBubble
      v-for="(m, i) in messages"
      :key="m.messageId ?? i"
      :message="m"
    />
  </div>
</template>

<style scoped>
.chat-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 20px 24px;
  width: 100%;
}
</style>
