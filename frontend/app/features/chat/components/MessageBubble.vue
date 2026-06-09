<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "~/composables/useI18n";

import type { ChatMessage } from "../composables/useChatStream";
import MessageContent from "./MessageContent.vue";
import SourcesPanel from "./SourcesPanel.vue";
import ToolTrace from "./ToolTrace.vue";

const props = defineProps<{
  message: ChatMessage;
}>();

const { t } = useI18n();

const isAssistant = computed(() => props.message.role === "assistant");
</script>

<template>
  <div :class="['chat-bubble', `chat-bubble--${message.role}`]">
    <div class="chat-bubble__role">
      {{ message.role === "user" ? t("chat.youLabel") : t("chat.assistantLabel") }}
    </div>

    <div class="chat-bubble__content">
      <MessageContent :content="message.content" :pending="message.pending" />
    </div>

    <!-- Assistant-only: live tool trace + expandable provenance. -->
    <ToolTrace
      v-if="isAssistant"
      :tool-calls="message.toolCalls"
      :validation="message.validation"
    />

    <!-- Blocked answers: show exactly which figures could not be verified. -->
    <div
      v-if="isAssistant && message.validation === 'blocked' && (message.unverified?.length ?? 0) > 0"
      class="chat-bubble__unverified"
    >
      <span class="chat-bubble__unverified-label">{{ t("chat.validation.unverifiedLabel") }}</span>
      <span
        v-for="(fig, i) in message.unverified"
        :key="`uv-${i}`"
        class="chat-bubble__unverified-chip"
        >{{ fig }}</span
      >
    </div>

    <SourcesPanel
      v-if="isAssistant"
      :sources="message.sources"
      :bindings="message.bindings"
    />
  </div>
</template>

<style scoped>
.chat-bubble {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 16px;
  border-radius: 12px;
  max-width: 720px;
  white-space: normal;
  word-break: break-word;
  line-height: 1.55;
  font-size: 14px;
}
.chat-bubble--user {
  align-self: flex-end;
  background: var(--bg-accent-subtle, #eef2ff);
  color: var(--text-primary, #0f172a);
  border: 1px solid var(--border-subtle, #e2e8f0);
}
.chat-bubble--assistant {
  align-self: flex-start;
  background: var(--bg-elevated, #ffffff);
  color: var(--text-primary, #0f172a);
  border: 1px solid var(--border-subtle, #e2e8f0);
}
.chat-bubble__role {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-tertiary, #64748b);
}
.chat-bubble__content {
  display: inline;
}
.chat-bubble__caret {
  display: inline-block;
  margin-left: 2px;
  font-size: 12px;
  color: var(--text-tertiary, #64748b);
  animation: chat-caret-blink 1s steps(2) infinite;
}
.chat-bubble__unverified {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  background: var(--bg-danger-subtle, #fef2f2);
  border: 1px solid var(--border-danger, #fecaca);
}
.chat-bubble__unverified-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-danger, #b91c1c);
}
.chat-bubble__unverified-chip {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 6px;
  background: var(--bg-elevated, #ffffff);
  border: 1px solid var(--border-danger, #fecaca);
  color: var(--text-danger, #b91c1c);
}
@keyframes chat-caret-blink {
  to {
    opacity: 0;
  }
}
</style>
