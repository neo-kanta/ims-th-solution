<script setup lang="ts">
definePageMeta({
  layout: 'dashboard',
  middleware: ['auth', 'permission'],
  meta: { permission: 'WATCHLIST_VIEW' },
})

const showAddDrawer   = ref(false)
const showRuleDrawer  = ref(false)
const selectedItem    = ref<string | null>('WL-001')

const watchlist = [
  { id: 'WL-001', security: 'CPALL TB',   type: 'Equity', lastNav: 49.10, prevNav: 48.80, change: 0.61, threshold: 52.00, direction: 'ABOVE', triggered: false, stale: false, lastAlert: '2026-03-11 09:15', cooldown: false },
  { id: 'WL-002', security: 'PTT TB',     type: 'Equity', lastNav: 69.00, prevNav: 68.50, change: 0.73, threshold: 65.00, direction: 'BELOW', triggered: false, stale: false, lastAlert: null, cooldown: false },
  { id: 'WL-003', security: 'ADVANC TB',  type: 'Equity', lastNav: 240.5, prevNav: 238.0, change: 1.05, threshold: 250.0, direction: 'ABOVE', triggered: false, stale: false, lastAlert: '2026-03-12 14:00', cooldown: true },
  { id: 'WL-004', security: 'KBANK TB',   type: 'Equity', lastNav: 142.0, prevNav: 144.5, change: -1.73, threshold: 140.0, direction: 'BELOW', triggered: true,  stale: false, lastAlert: '2026-03-13 10:30', cooldown: false },
  { id: 'WL-005', security: 'AOT TB',     type: 'Equity', lastNav: 60.5,  prevNav: 61.0,  change: -0.82, threshold: 58.0,  direction: 'BELOW', triggered: false, stale: true,  lastAlert: null, cooldown: false },
  { id: 'WL-006', security: 'GOVT10Y',    type: 'Bonds',  lastNav: 10020, prevNav: 10000, change: 0.20, threshold: 10100, direction: 'ABOVE', triggered: false, stale: false, lastAlert: null, cooldown: false },
]

const selected = computed(() => watchlist.find(w => w.id === selectedItem.value))

const distancePct = (w: typeof watchlist[0]) =>
  ((w.threshold - w.lastNav) / w.lastNav * 100)
</script>

<template>
  <div>
    <AppPageHeader
      title="Watchlist"
      description="Monitor selected instruments with threshold-based alert rules"
    >
      <template #actions>
        <button class="btn btn-secondary btn-sm" @click="showRuleDrawer=true">Manage Rules</button>
        <button class="btn btn-primary btn-sm" @click="showAddDrawer=true">
          <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 3v10M3 8h10"/></svg>
          Add to Watchlist
        </button>
      </template>
    </AppPageHeader>

    <!-- Stats -->
    <div class="grid-4" style="margin-bottom:16px;">
      <div class="stat-card"><div class="stat-card-label">Watching</div><div class="stat-card-value">{{ watchlist.length }}</div><div class="stat-card-meta">Instruments</div></div>
      <div class="stat-card"><div class="stat-card-label">Active Alerts</div><div class="stat-card-value" style="color:#dc2626;">{{ watchlist.filter(w=>w.triggered).length }}</div><div class="stat-card-meta">Threshold crossed</div></div>
      <div class="stat-card"><div class="stat-card-label">On Cooldown</div><div class="stat-card-value" style="color:#d97706;">{{ watchlist.filter(w=>w.cooldown).length }}</div><div class="stat-card-meta">Suppressed alerts</div></div>
      <div class="stat-card"><div class="stat-card-label">Stale Data</div><div class="stat-card-value" style="color:#94a3b8;">{{ watchlist.filter(w=>w.stale).length }}</div><div class="stat-card-meta">NAV not refreshed</div></div>
    </div>

    <!-- Split Layout -->
    <div style="display:grid;grid-template-columns:1fr 360px;gap:16px;align-items:start;">
      <!-- Watchlist Table -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">Watchlist Items</span>
          <div style="display:flex;gap:6px;margin-left:auto;">
            <select class="form-select" style="height:28px;font-size:12px;width:110px;padding:0 8px;"><option>All Types</option><option>Equity</option><option>Bonds</option></select>
            <select class="form-select" style="height:28px;font-size:12px;width:120px;padding:0 8px;"><option>All Status</option><option>Triggered</option><option>Normal</option></select>
          </div>
        </div>
        <table class="data-table">
          <thead>
            <tr>
              <th>Security</th>
              <th>Type</th>
              <th class="col-right">Last NAV</th>
              <th class="col-right">Change %</th>
              <th>Threshold</th>
              <th class="col-right">Distance</th>
              <th>Status</th>
              <th>Last Alert</th>
              <th class="col-action"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="w in watchlist" :key="w.id"
              @click="selectedItem = w.id"
              style="cursor:pointer;"
              :class="[selectedItem===w.id ? 'selected' : '', w.triggered ? 'row-error' : '']"
            >
              <td>
                <div style="font-weight:600;">{{ w.security }}</div>
                <div v-if="w.stale" style="font-size:10px;color:#d97706;">⚠ Stale NAV</div>
              </td>
              <td><span class="badge badge-neutral" style="font-size:10px;">{{ w.type }}</span></td>
              <td class="col-right numeric" :style="w.stale ? 'color:#94a3b8;' : ''">฿{{ w.lastNav.toLocaleString() }}</td>
              <td class="col-right numeric" :class="w.change >= 0 ? 'positive' : 'negative'">
                {{ w.change >= 0 ? '+' : '' }}{{ w.change.toFixed(2) }}%
              </td>
              <td>
                <div style="display:flex;align-items:center;gap:4px;font-size:12.5px;">
                  <span class="badge" :class="w.direction==='ABOVE'?'badge-info':'badge-warning'" style="font-size:10px;">{{ w.direction }}</span>
                  <span class="numeric">฿{{ w.threshold.toLocaleString() }}</span>
                </div>
              </td>
              <td class="col-right">
                <span class="numeric" style="font-size:12.5px;" :style="Math.abs(distancePct(w)) < 3 ? 'color:#d97706;font-weight:600;' : 'color:#64748b;'">
                  {{ distancePct(w) > 0 ? '+' : '' }}{{ distancePct(w).toFixed(1) }}%
                </span>
              </td>
              <td>
                <span v-if="w.triggered"  class="badge badge-error"   style="font-size:10px;">TRIGGERED</span>
                <span v-else-if="w.cooldown" class="badge badge-warning" style="font-size:10px;">COOLDOWN</span>
                <span v-else-if="w.stale"   class="badge badge-neutral" style="font-size:10px;">STALE</span>
                <span v-else                class="badge badge-success" style="font-size:10px;">MONITORING</span>
              </td>
              <td class="col-mono" style="font-size:11.5px;color:#64748b;">{{ w.lastAlert ?? '—' }}</td>
              <td class="col-action">
                <button class="btn btn-ghost btn-icon-sm">
                  <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="8" cy="8" r="1.5"/><circle cx="3" cy="8" r="1.5"/><circle cx="13" cy="8" r="1.5"/></svg>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <div class="card-footer" style="display:flex;justify-content:space-between;align-items:center;">
          <span style="font-size:12px;color:#64748b;">{{ watchlist.length }} items</span>
          <span style="font-size:11.5px;color:#94a3b8;">Last NAV refresh: 13 Mar 2026 14:30 ICT</span>
        </div>
      </div>

      <!-- Detail Panel -->
      <div v-if="selected" style="display:flex;flex-direction:column;gap:12px;">
        <div class="card">
          <div class="card-header">
            <div><div class="card-title">{{ selected.security }}</div><div class="card-subtitle">{{ selected.type }} &nbsp;·&nbsp; {{ selected.id }}</div></div>
            <span v-if="selected.triggered" class="badge badge-error">TRIGGERED</span>
            <span v-else-if="selected.stale" class="badge badge-neutral">STALE NAV</span>
            <span v-else class="badge badge-success">MONITORING</span>
          </div>
          <div class="card-body" style="display:flex;flex-direction:column;gap:12px;">
            <!-- Price gauge -->
            <div style="padding:12px;background:#f8fafc;border-radius:6px;border:1px solid #e2e8f0;">
              <div style="display:flex;justify-content:space-between;margin-bottom:6px;font-size:12px;color:#64748b;">
                <span>Last NAV</span><span>Threshold ({{ selected.direction }})</span>
              </div>
              <div style="display:flex;justify-content:space-between;align-items:baseline;margin-bottom:8px;">
                <div class="numeric" style="font-size:20px;font-weight:700;" :style="selected.stale?'color:#94a3b8':''">฿{{ selected.lastNav.toLocaleString() }}</div>
                <div class="numeric" style="font-size:16px;font-weight:600;color:#374151;">฿{{ selected.threshold.toLocaleString() }}</div>
              </div>
              <!-- Progress bar showing distance to threshold -->
              <div style="height:6px;background:#e2e8f0;border-radius:3px;overflow:hidden;margin-bottom:4px;">
                <div :style="`width:${Math.min(Math.abs(distancePct(selected))*10,100)}%;background:${selected.triggered ? '#dc2626' : Math.abs(distancePct(selected)) < 3 ? '#d97706' : '#2563eb'};height:100%;border-radius:3px;`"></div>
              </div>
              <div style="font-size:11.5px;color:#64748b;text-align:center;">
                {{ Math.abs(distancePct(selected)).toFixed(1) }}% {{ selected.direction === 'ABOVE' ? 'below' : 'above' }} threshold
              </div>
            </div>

            <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px;">
              <div><div class="text-label" style="margin-bottom:2px;">Change</div><div class="numeric" :class="selected.change>=0?'positive':'negative'" style="font-size:13px;font-weight:600;">{{ selected.change>=0?'+':'' }}{{ selected.change.toFixed(2) }}%</div></div>
              <div><div class="text-label" style="margin-bottom:2px;">Prev NAV</div><div class="numeric" style="font-size:13px;">฿{{ selected.prevNav.toLocaleString() }}</div></div>
              <div><div class="text-label" style="margin-bottom:2px;">Alert Direction</div><span class="badge" :class="selected.direction==='ABOVE'?'badge-info':'badge-warning'">{{ selected.direction }}</span></div>
              <div><div class="text-label" style="margin-bottom:2px;">Cooldown</div><span class="badge" :class="selected.cooldown?'badge-warning':'badge-neutral'">{{ selected.cooldown ? 'Active' : 'None' }}</span></div>
            </div>

            <div v-if="selected.stale" class="alert alert-warning">
              <svg width="13" height="13" viewBox="0 0 16 16" fill="currentColor" style="flex-shrink:0;"><path d="M8 2L1 14h14L8 2z"/></svg>
              <div style="font-size:12px;">NAV data is stale (&gt;4 hours). Threshold comparison may be inaccurate. Trigger NAV refresh to update.</div>
            </div>

            <div v-if="selected.lastAlert" style="padding:8px 10px;background:#f8fafc;border-radius:5px;border:1px solid #e2e8f0;">
              <div class="text-label" style="margin-bottom:2px;">Last Alert Sent</div>
              <div class="col-mono" style="font-size:12.5px;color:#475569;">{{ selected.lastAlert }}</div>
            </div>
          </div>
          <div class="card-footer" style="display:flex;gap:6px;justify-content:flex-end;">
            <button class="btn btn-secondary btn-sm">Edit Rule</button>
            <button class="btn btn-danger btn-sm">Remove</button>
            <button class="btn btn-primary btn-sm" @click="showRuleDrawer=true">Adjust Threshold</button>
          </div>
        </div>

        <!-- Alert History -->
        <div class="card">
          <div class="card-header"><span class="card-title">Recent Alerts</span></div>
          <div class="card-body" style="display:flex;flex-direction:column;gap:6px;">
            <div v-if="selected.lastAlert"
              style="display:flex;align-items:flex-start;gap:8px;padding:8px 10px;background:#fff5f5;border-radius:5px;border:1px solid #fecaca;">
              <svg width="13" height="13" viewBox="0 0 16 16" fill="#dc2626" style="flex-shrink:0;margin-top:1px;"><path d="M8 1l7 13H1L8 1z"/></svg>
              <div>
                <div style="font-size:12.5px;font-weight:600;color:#991b1b;">Threshold Alert</div>
                <div style="font-size:11.5px;color:#64748b;">{{ selected.security }} {{ selected.direction === 'ABOVE' ? 'crossed above' : 'fell below' }} ฿{{ selected.threshold }}</div>
                <div class="col-mono" style="font-size:11px;color:#94a3b8;">{{ selected.lastAlert }}</div>
              </div>
            </div>
            <div v-else class="empty-state" style="padding:16px;">
              <span class="empty-desc">No alerts triggered yet</span>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="card">
        <div class="empty-state">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#cbd5e1" stroke-width="1.5"><path d="M18 8A6 6 0 006 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 01-3.46 0"/></svg>
          <div class="empty-title">Select a watchlist item</div>
          <div class="empty-desc">Click any row to view price details and alert rules</div>
        </div>
      </div>
    </div>

    <!-- Add to Watchlist Drawer -->
    <template v-if="showAddDrawer">
      <div class="drawer-overlay" @click="showAddDrawer=false"></div>
      <div class="drawer">
        <div class="drawer-header">
          <div><div class="drawer-title">Add to Watchlist</div><div class="drawer-subtitle">Set threshold alert rules for an instrument</div></div>
          <button class="btn btn-ghost btn-icon-sm" @click="showAddDrawer=false"><svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 3l10 10M13 3L3 13"/></svg></button>
        </div>
        <div class="drawer-body" style="display:flex;flex-direction:column;gap:14px;">
          <div class="form-group"><label class="form-label form-label-req">Security / Instrument</label><input class="form-input" placeholder="e.g. CPALL TB, PTT TB, GOVT10Y" /></div>
          <div class="form-group"><label class="form-label form-label-req">Instrument Type</label>
            <select class="form-select"><option>Equity</option><option>Bonds</option><option>Mutual Fund</option><option>ETF</option></select>
          </div>
          <hr class="divider" />
          <div style="font-size:13px;font-weight:600;color:#374151;">Alert Threshold Rule</div>
          <div class="form-grid-2">
            <div class="form-group"><label class="form-label form-label-req">Trigger Direction</label>
              <select class="form-select"><option>ABOVE — alert when NAV exceeds threshold</option><option>BELOW — alert when NAV falls below threshold</option></select>
            </div>
            <div class="form-group"><label class="form-label form-label-req">Threshold Value (฿)</label><input class="form-input numeric" type="number" placeholder="0.00" /></div>
          </div>
          <div class="form-group"><label class="form-label">Alert Cooldown (hours)</label>
            <input class="form-input numeric" type="number" value="4" min="1" max="24" />
            <div class="form-helper">Minimum hours between repeated alerts for the same instrument</div>
          </div>
          <div class="form-group"><label class="form-label">Notification Channels</label>
            <div style="display:flex;flex-direction:column;gap:6px;">
              <label style="display:flex;align-items:center;gap:6px;cursor:pointer;font-size:13px;"><input type="checkbox" checked /> In-app notification</label>
              <label style="display:flex;align-items:center;gap:6px;cursor:pointer;font-size:13px;"><input type="checkbox" /> Email (SMTP placeholder)</label>
            </div>
          </div>
        </div>
        <div class="drawer-footer">
          <button class="btn btn-secondary btn-sm" @click="showAddDrawer=false">Cancel</button>
          <button class="btn btn-primary btn-sm" @click="showAddDrawer=false">Add to Watchlist</button>
        </div>
      </div>
    </template>
  </div>
</template>
