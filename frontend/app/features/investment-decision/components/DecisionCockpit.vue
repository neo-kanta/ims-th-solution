<template>
  <div class="legacy-screen">
    <div class="legacy-toolbar">
      <div class="legacy-toolbar__title">
        <slot name="title" />
      </div>
      <div class="legacy-toolbar__buttons">
        <slot name="toolbar" />
      </div>
    </div>
    <div class="legacy-layout">
      <div class="legacy-col legacy-col--left">
        <slot name="left" />
      </div>
      <div class="legacy-col legacy-col--middle">
        <slot name="middle" />
      </div>
      <div class="legacy-col legacy-col--right">
        <slot name="right" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.legacy-screen {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  overflow: hidden;
  color: var(--text-primary);
  box-shadow: var(--shadow-md);
  transition: background-color 0.2s, border-color 0.2s;
}

.legacy-toolbar {
  background: var(--bg-card-muted, #f6f8fa);
  border-bottom: 1px solid var(--border-subtle);
  padding: 10px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.legacy-toolbar__title {
  font-size: 13px;
  font-weight: 700;
  color: var(--state-info, #0969da);
  font-family: monospace;
}

.legacy-toolbar__buttons {
  display: flex;
  gap: 8px;
}

.legacy-layout {
  display: flex;
  min-height: 520px;
  background: var(--bg-card);
}

.legacy-col {
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--border-subtle);
}

.legacy-col:last-child {
  border-right: none;
}

.legacy-col--left {
  width: 25%;
  min-width: 200px;
  background: var(--bg-card-muted);
}

.legacy-col--middle {
  width: 55%;
  flex-grow: 1;
}

.legacy-col--right {
  width: 20%;
  min-width: 180px;
  background: var(--bg-card-muted);
}
</style>
