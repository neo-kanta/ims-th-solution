<script setup lang="ts">
import { computed, ref } from "vue";
import FinancialVisualizer from "./FinancialVisualizer.vue";

const props = defineProps<{
  content: string;
  pending?: boolean;
}>();

interface MarkdownNode {
  type: "text" | "heading" | "list" | "code" | "table";
  level?: number;
  ordered?: boolean;
  items?: string[];
  lang?: string;
  content?: string;
  headers?: string[];
  rows?: string[][];
}

// Custom Markdown block parser
const nodes = computed<MarkdownNode[]>(() => {
  const list: MarkdownNode[] = [];
  const lines = props.content.split(/\r?\n/);
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    // 1. Code Blocks
    if (line.trim().startsWith("```")) {
      const lang = line.trim().slice(3).trim();
      let code = "";
      i++;
      while (i < lines.length && !lines[i].trim().startsWith("```")) {
        code += lines[i] + "\n";
        i++;
      }
      list.push({ type: "code", lang: lang || "text", content: code.trim() });
      i++;
      continue;
    }

    // 2. Tables
    // A table starts with a row like | col1 | col2 |
    // and is followed by a separator row like | --- | --- |
    if (
      line.trim().startsWith("|") &&
      i + 1 < lines.length &&
      lines[i + 1].trim().match(/^\|\s*[-:]+[-:\s|]*\|$/)
    ) {
      const headers = line
        .split("|")
        .map((s) => s.trim())
        .filter((_, idx, arr) => idx > 0 && idx < arr.length - 1);
      
      const rows: string[][] = [];
      i += 2; // skip header and separator
      
      while (i < lines.length && lines[i].trim().startsWith("|")) {
        const row = lines[i]
          .split("|")
          .map((s) => s.trim())
          .filter((_, idx, arr) => idx > 0 && idx < arr.length - 1);
        rows.push(row);
        i++;
      }
      list.push({ type: "table", headers, rows });
      continue;
    }

    // 3. Headings
    if (line.trim().startsWith("#")) {
      const match = line.trim().match(/^(#{1,6})\s+(.*)$/);
      if (match) {
        list.push({
          type: "heading",
          level: match[1].length,
          content: match[2],
        });
        i++;
        continue;
      }
    }

    // 4. Unordered Lists
    if (line.trim().startsWith("- ") || line.trim().startsWith("* ")) {
      const items: string[] = [];
      while (
        i < lines.length &&
        (lines[i].trim().startsWith("- ") || lines[i].trim().startsWith("* "))
      ) {
        items.push(lines[i].trim().slice(2).trim());
        i++;
      }
      list.push({ type: "list", ordered: false, items });
      continue;
    }

    // 5. Ordered Lists
    if (line.trim().match(/^\d+\.\s/)) {
      const items: string[] = [];
      while (i < lines.length && lines[i].trim().match(/^\d+\.\s/)) {
        const match = lines[i].trim().match(/^\d+\.\s+(.*)$/);
        if (match) {
          items.push(match[1].trim());
        }
        i++;
      }
      list.push({ type: "list", ordered: true, items });
      continue;
    }

    // 6. Paragraphs (standard text)
    if (line.trim() === "") {
      i++;
      continue;
    }

    let paragraphContent = "";
    while (
      i < lines.length &&
      lines[i].trim() !== "" &&
      !lines[i].trim().startsWith("```") &&
      !lines[i].trim().startsWith("#") &&
      !(
        lines[i].trim().startsWith("|") &&
        i + 1 < lines.length &&
        lines[i + 1].trim().match(/^\|\s*[-:]+[-:\s|]*\|$/)
      ) &&
      !lines[i].trim().startsWith("- ") &&
      !lines[i].trim().startsWith("* ") &&
      !lines[i].trim().match(/^\d+\.\s/)
    ) {
      paragraphContent += (paragraphContent ? "\n" : "") + lines[i];
      i++;
    }
    list.push({ type: "text", content: paragraphContent });
  }

  return list;
});

// Inline formatting (bold, italic, inline code, links)
function formatInline(text: string): string {
  // Escape HTML tags to prevent XSS/rendering issues
  let html = text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");

  // Bold: **text**
  html = html.replace(/\*\*(.*?)\*\*/g, "<strong>$1</strong>");

  // Italics: *text*
  html = html.replace(/\*(.*?)\*/g, "<em>$1</em>");

  // Inline Code: `code`
  html = html.replace(/`(.*?)`/g, "<code>$1</code>");

  // Links: [label](url)
  html = html.replace(
    /\[(.*?)\]\((.*?)\)/g,
    '<a href="$2" target="_blank" class="rich-content__link">$1</a>',
  );

  return html;
}

// Copy button reactive state per code block
const copiedStates = ref<Record<number, boolean>>({});

function copyCode(code: string, idx: number) {
  navigator.clipboard.writeText(code).then(() => {
    copiedStates.value[idx] = true;
    setTimeout(() => {
      copiedStates.value[idx] = false;
    }, 2000);
  });
}

const isThinking = computed(() => {
  return props.pending && (!props.content || props.content.trim() === "");
});
</script>

<template>
  <div class="rich-content">
    <!-- Thinking dots animation -->
    <div v-if="isThinking" class="rich-content__thinking">
      <span class="rich-content__thinking-dot"></span>
      <span class="rich-content__thinking-dot"></span>
      <span class="rich-content__thinking-dot"></span>
    </div>

    <template v-else>
      <div v-for="(node, idx) in nodes" :key="idx" class="rich-content__node">
      <!-- Heading -->
      <component
        :is="'h' + node.level"
        v-if="node.type === 'heading'"
        :class="['rich-content__heading', `rich-content__h${node.level}`]"
        v-html="formatInline(node.content || '')"
      />

      <!-- List -->
      <component
        :is="node.ordered ? 'ol' : 'ul'"
        v-else-if="node.type === 'list'"
        :class="[
          'rich-content__list',
          node.ordered ? 'rich-content__list--ordered' : 'rich-content__list--unordered',
        ]"
      >
        <li
          v-for="(item, itemIdx) in node.items"
          :key="itemIdx"
          class="rich-content__list-item"
          v-html="formatInline(item)"
        />
      </component>

      <!-- Code Block -->
      <div v-else-if="node.type === 'code'" class="rich-content__code-block">
        <div class="rich-content__code-header">
          <span class="rich-content__code-lang">{{ node.lang }}</span>
          <button
            type="button"
            class="rich-content__code-copy"
            @click="copyCode(node.content || '', idx)"
          >
            {{ copiedStates[idx] ? "Copied!" : "Copy" }}
          </button>
        </div>
        <pre class="rich-content__code-pre"><code>{{ node.content }}</code></pre>
      </div>

      <!-- Table / Chart Visualizer -->
      <FinancialVisualizer
        v-else-if="node.type === 'table'"
        :headers="node.headers || []"
        :rows="node.rows || []"
      />

      <!-- Plain Paragraph -->
      <p
        v-else
        class="rich-content__paragraph"
      >
        <span v-html="formatInline(node.content || '')" /><span
          v-if="pending && idx === nodes.length - 1"
          class="rich-content__caret"
          aria-hidden="true"
        >&#9608;</span>
      </p>
    </div>

    <!-- Caret fallback when the last node is a table, list, or code block -->
    <div
      v-if="pending && nodes.length > 0 && !['text', 'heading'].includes(nodes[nodes.length - 1].type)"
      class="rich-content__caret-fallback"
    >
      <span class="rich-content__caret" aria-hidden="true">&#9608;</span>
    </div>
    </template>
  </div>
</template>

<style>
/* Scoped styles would not apply to v-html elements easily, so we use unscoped
   nested classes or standard target class styling. */
.rich-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.rich-content__heading {
  margin: 16px 0 8px;
  font-weight: 700;
  color: var(--text-primary, #0f172a);
  line-height: 1.3;
}

.rich-content__h1 { font-size: 1.6em; }
.rich-content__h2 { font-size: 1.4em; }
.rich-content__h3 { font-size: 1.2em; border-bottom: 1px solid var(--border-subtle); padding-bottom: 4px; }
.rich-content__h4 { font-size: 1.1em; }
.rich-content__h5 { font-size: 1em; }
.rich-content__h6 { font-size: 0.9em; }

.rich-content__paragraph {
  margin: 0;
  line-height: 1.6;
  color: var(--text-secondary, #334155);
}

.rich-content__paragraph strong,
.rich-content__list-item strong {
  font-weight: 600;
  color: var(--text-primary, #0f172a);
}

.rich-content__paragraph code,
.rich-content__list-item code {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 0.9em;
  background: var(--bg-canvas, #f1f5f9);
  padding: 2px 6px;
  border-radius: 4px;
  color: var(--text-accent, #4338ca);
  border: 1px solid var(--border-subtle, #e2e8f0);
}

.rich-content__list {
  margin: 4px 0 8px;
  padding-left: 24px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.rich-content__list--unordered {
  list-style-type: disc;
}

.rich-content__list--ordered {
  list-style-type: decimal;
}

.rich-content__list-item {
  line-height: 1.5;
  color: var(--text-secondary, #334155);
}

.rich-content__link {
  color: var(--action-primary, #3b82f6);
  text-decoration: underline;
  font-weight: 500;
  transition: color 0.15s ease;
}

.rich-content__link:hover {
  color: var(--action-primary-hover, #2563eb);
}

/* Code block styling */
.rich-content__code-block {
  border: 1px solid var(--border-subtle, #e2e8f0);
  border-radius: 8px;
  overflow: hidden;
  background: var(--bg-canvas, #1e293b);
  margin: 12px 0;
}

.rich-content__code-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 12px;
  background: rgba(15, 23, 42, 0.6);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.rich-content__code-lang {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 11px;
  text-transform: uppercase;
  color: #94a3b8;
  font-weight: 600;
}

.rich-content__code-copy {
  background: transparent;
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #e2e8f0;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.rich-content__code-copy:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #ffffff;
  border-color: rgba(255, 255, 255, 0.4);
}

.rich-content__code-pre {
  margin: 0;
  padding: 12px;
  overflow-x: auto;
}

.rich-content__code-pre code {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 13px;
  color: #f8fafc;
  line-height: 1.5;
  white-space: pre;
}

.rich-content__caret {
  display: inline-block;
  margin-left: 2px;
  font-size: 12px;
  color: var(--text-tertiary, #64748b);
  animation: rich-caret-blink 1s steps(2) infinite;
}

.rich-content__caret-fallback {
  margin-top: 4px;
}

.rich-content__thinking {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 6px 4px;
}

.rich-content__thinking-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background-color: var(--text-tertiary, #64748b);
  opacity: 0.4;
  animation: rich-thinking-bounce 1.4s infinite ease-in-out both;
}

.rich-content__thinking-dot:nth-child(1) {
  animation-delay: -0.32s;
}

.rich-content__thinking-dot:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes rich-thinking-bounce {
  0%, 80%, 100% {
    transform: scale(0.3);
    opacity: 0.2;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

@keyframes rich-caret-blink {
  to {
    opacity: 0;
  }
}
</style>
