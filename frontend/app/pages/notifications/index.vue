<script setup lang="ts">
definePageMeta({
  layout: 'dashboard',
  middleware: ['auth'],
})

// Redirect to the notification settings / center page
// This page serves as the user-facing notification inbox accessible from the topbar bell icon.

const selectedId     = ref<string | null>(null)
const filterSeverity = ref('All')
const filterRead     = ref('All')

const severityMap: Record<string, { label: string; badge: string; color: string }> = {
  CRITICAL: { label: 'Critical', badge: 'badge-error',   color: '#dc2626' },
  HIGH:     { label: 'High',     badge: 'badge-warning', color: '#d97706' },
  MEDIUM:   { label: 'Medium',   badge: 'badge-info',    color: '#0284c7' },
  LOW:      { label: 'Low',      badge: 'badge-neutral', color: '#64748b' },
}

const notifications = [
  { id: 'N-0051', ts: '2026-03-13 11:06:01', severity: 'CRITICAL', title: 'IRG BLOCK — DEC-2026-0041 rejected', body: 'Decision DEC-2026-0041 for CPALL TB has been blocked by IRG rule CATEGORY_LIMIT. The equity exposure limit (75%) would be exceeded. Immediate review required.', module: 'Investment', read: false, actionUrl: '/investment/decision' },
  { id: 'N-0050', ts: '2026-03-13 10:45:00', severity: 'HIGH',     title: 'Approval required — DEC-2026-0041', body: 'Investment decision DEC-2026-0041 is pending your approval. IRG yielded 1 warning. Review and approve or reject before end of trading day.', module: 'Investment', read: false, actionUrl: '/approval' },
  { id: 'N-0049', ts: '2026-03-13 10:31:00', severity: 'MEDIUM',   title: 'New research report submitted', body: 'apinya.r submitted research report RES-2026-0012 for KBANK TB. Status: PENDING. Available for review.', module: 'Research', read: true, actionUrl: '/investment/analysis' },
  { id: 'N-0048', ts: '2026-03-13 08:02:30', severity: 'LOW',      title: 'Day Start completed — ABCFLEX1', body: 'Investment day for contract ABCFLEX1 has been started by somchai.w. 2 carry-forward items. Workflow stage: OPEN.', module: 'Workflow', read: true, actionUrl: '/workflow' },
  { id: 'N-0047', ts: '2026-03-13 07:58:00', severity: 'MEDIUM',   title: 'NAV data stale — ABCBOND2', body: 'Portfolio NAV for ABCBOND2 has not been refreshed for 26 hours. Market data integration may be experiencing issues.', module: 'Portfolio', read: true, actionUrl: '/portfolio' },
  { id: 'N-0046', ts: '2026-03-12 17:46:00', severity: 'HIGH',     title: 'Permission granted — apinya.r', body: 'Administrator admin.sys granted permission RESEARCH_SUBMIT to apinya.r. Takes effect at next login.', module: 'Permissions', read: true, actionUrl: '/permissions/accounts' },
  { id: 'N-0045', ts: '2026-03-12 15:00:00', severity: 'LOW',      title: 'Leave request submitted — wichai.p', body: 'wichai.p submitted temporary leave LV-2026-0019 for 2026-03-13. Pending approval. No delegation agents configured.', module: 'Leave', read: true, actionUrl: '/leave' },
]

const filtered = computed(() => notifications.filter(n => {
  if (filterSeverity.value !== 'All' && n.severity !== filterSeverity.value) return false
  if (filterRead.value === 'Unread' && n.read) return false
  if (filterRead.value === 'Read' && !n.read) return false
  return true
}))

const selected   = computed(() => notifications.find(n => n.id === selectedId.value))
const unreadCount = computed(() => notifications.filter(n => !n.read).length)
</script>

<template>
  <div>
    <AppPageHeader
      title="My Notifications"
      description="System alerts and events routed to you based on your role and subscriptions"
    >
      <template #eyebrow>
        <div class="breadcrumb" style="margin-bottom:4px;"><span>Notifications</span></div>
      </template>
      <template #actions>
        <div style="display:flex;gap:8px;align-items:center;">
          <span v-if="unreadCount>0" class="badge badge-error">{{ unreadCount }} unread</span>
          <button class="btn btn-secondary btn-sm">Mark All Read</button>
          <NuxtLink to="/settings/notifications" class="btn btn-secondary btn-sm">Notification Settings</NuxtLink>
        </div>
      </template>
    </AppPageHeader>

    <div class="grid-4" style="margin-bottom:16px;">
      <div class="stat-card"><div class="stat-card-label">Total</div><div class="stat-card-value">{{ notifications.length }}</div><div class="stat-card-meta">All time today</div></div>
      <div class="stat-card"><div class="stat-card-label">Unread</div><div class="stat-card-value" style="color:#dc2626;">{{ unreadCount }}</div><div class="stat-card-meta">Require attention</div></div>
      <div class="stat-card"><div class="stat-card-label">Critical / High</div><div class="stat-card-value" style="color:#d97706;">{{ notifications.filter(n=>n.severity==='CRITICAL'||n.severity==='HIGH').length }}</div><div class="stat-card-meta">High priority</div></div>
      <div class="stat-card"><div class="stat-card-label">Modules</div><div class="stat-card-value">{{ new Set(notifications.map(n=>n.module)).size }}</div><div class="stat-card-meta">Event sources</div></div>
    </div>

    <div class="card" style="margin-bottom:14px;">
      <div class="card-body" style="padding:10px 16px;">
        <div class="filter-bar">
          <span style="font-size:12px;font-weight:500;color:#374151;">Filter:</span>
          <select v-model="filterSeverity" class="form-select" style="height:28px;font-size:12px;width:130px;padding:0 8px;">
            <option>All</option>
            <option>CRITICAL</option>
            <option>HIGH</option>
            <option>MEDIUM</option>
            <option>LOW</option>
          </select>
          <select v-model="filterRead" class="form-select" style="height:28px;font-size:12px;width:110px;padding:0 8px;">
            <option>All</option>
            <option>Unread</option>
            <option>Read</option>
          </select>
          <select class="form-select" style="height:28px;font-size:12px;width:140px;padding:0 8px;">
            <option>All Modules</option>
            <option>Investment</option>
            <option>Workflow</option>
            <option>Research</option>
            <option>Portfolio</option>
            <option>Permissions</option>
            <option>Leave</option>
          </select>
          <button class="btn btn-secondary btn-sm">Reset</button>
        </div>
      </div>
    </div>

    <div style="display:grid;grid-template-columns:1fr 420px;gap:16px;align-items:start;">
      <!-- Notification list -->
      <div class="card" style="padding:0;overflow:hidden;">
        <div v-for="n in filtered" :key="n.id"
          @click="selectedId=n.id"
          style="display:flex;gap:12px;padding:14px 16px;border-bottom:1px solid #f1f5f9;cursor:pointer;transition:background 0.08s;"
          :style="selectedId===n.id ? 'background:#eff6ff;' : !n.read ? 'background:#fffbeb;' : ''"
        >
          <div style="flex-shrink:0;display:flex;flex-direction:column;align-items:center;gap:6px;padding-top:2px;">
            <span style="width:8px;height:8px;border-radius:50%;flex-shrink:0;" :style="`background:${!n.read ? severityMap[n.severity]?.color : '#cbd5e1'};`"></span>
          </div>
          <div style="flex:1;min-width:0;">
            <div style="display:flex;align-items:center;gap:8px;margin-bottom:3px;">
              <span class="badge" :class="severityMap[n.severity]?.badge" style="font-size:9px;padding:1px 5px;">{{ n.severity }}</span>
              <span class="badge badge-neutral" style="font-size:9px;padding:1px 5px;">{{ n.module }}</span>
              <span style="margin-left:auto;font-size:11px;color:#94a3b8;white-space:nowrap;">{{ n.ts.slice(11, 16) }}</span>
            </div>
            <div style="font-size:13px;font-weight:600;color:#1e293b;margin-bottom:3px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;" :style="n.read?'font-weight:500;color:#374151;':''">{{ n.title }}</div>
            <div style="font-size:12px;color:#64748b;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">{{ n.body }}</div>
          </div>
        </div>
        <div v-if="!filtered.length" class="empty-state" style="padding:48px;">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#cbd5e1" stroke-width="1.5"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
          <div class="empty-title">No notifications</div>
          <div class="empty-desc">No notifications match the selected filters</div>
        </div>
      </div>

      <!-- Detail panel -->
      <div v-if="selected" style="display:flex;flex-direction:column;gap:12px;">
        <div class="card">
          <div class="card-header">
            <div>
              <div style="display:flex;align-items:center;gap:6px;margin-bottom:4px;">
                <span class="badge" :class="severityMap[selected.severity]?.badge">{{ severityMap[selected.severity]?.label }}</span>
                <span class="badge badge-neutral" style="font-size:10px;">{{ selected.module }}</span>
              </div>
              <div class="card-title" style="font-size:14px;">{{ selected.title }}</div>
            </div>
          </div>
          <div class="card-body" style="display:flex;flex-direction:column;gap:10px;">
            <div style="padding:12px 14px;background:#f8fafc;border-radius:6px;border:1px solid #e2e8f0;">
              <div style="font-size:13px;line-height:1.75;color:#374151;">{{ selected.body }}</div>
            </div>
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px;">
              <div><div class="text-label" style="margin-bottom:2px;">Received</div><div class="col-mono" style="font-size:12px;color:#475569;">{{ selected.ts }}</div></div>
              <div><div class="text-label" style="margin-bottom:2px;">Notification ID</div><div class="col-mono" style="font-size:12px;color:#94a3b8;">{{ selected.id }}</div></div>
            </div>
            <div style="display:flex;gap:6px;">
              <button class="btn btn-secondary btn-sm" style="flex:1;">Dismiss</button>
              <NuxtLink :to="selected.actionUrl" class="btn btn-primary btn-sm" style="flex:1;text-align:center;">Go to Source</NuxtLink>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="card">
        <div class="empty-state" style="padding:48px 24px;">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#cbd5e1" stroke-width="1.5"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
          <div class="empty-title">Select a notification</div>
          <div class="empty-desc">Click a notification to read the full message and navigate to the source</div>
        </div>
      </div>
    </div>
  </div>
</template>
