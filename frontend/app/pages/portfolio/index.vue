<script setup lang="ts">
definePageMeta({
  layout: 'dashboard',
  middleware: ['auth', 'permission'],
  meta: { permission: 'PORTFOLIO_VIEW' },
})

const activeTab      = ref('holdings')
const selectedPos    = ref<string | null>(null)
const showPosDrawer  = ref(false)
const staleNAVWarn   = ref(true)
const filterContract = ref('ABCFLEX1')

const holdings = [
  { id: 'h1', security: 'CPALL TB',   category: 'Equity', sector: 'Commerce',     qty: 500000, avgCost: 45.20, lastNav: 49.10, mktValue: 24550000, unrealPL: 1950000, plPct: 8.63, weight: 23.4 },
  { id: 'h2', security: 'PTT TB',     category: 'Equity', sector: 'Energy',       qty: 200000, avgCost: 68.50, lastNav: 69.00, mktValue: 13800000, unrealPL: 100000,  plPct: 0.73, weight: 13.2 },
  { id: 'h3', security: 'ADVANC TB',  category: 'Equity', sector: 'Technology',   qty: 120000, avgCost: 235.0, lastNav: 240.5, mktValue: 28860000, unrealPL: 660000,  plPct: 2.34, weight: 27.5 },
  { id: 'h4', security: 'KBANK TB',   category: 'Equity', sector: 'Banking',      qty: 80000,  avgCost: 145.0, lastNav: 142.0, mktValue: 11360000, unrealPL: -240000, plPct: -2.07, weight: 10.8 },
  { id: 'h5', security: 'THB Cash',   category: 'Cash',   sector: '—',            qty: 1,      avgCost: 12450000, lastNav: 12450000, mktValue: 12450000, unrealPL: 0, plPct: 0, weight: 11.9 },
  { id: 'h6', security: 'GOVT10Y',    category: 'Bonds',  sector: 'Government',   qty: 1000,   avgCost: 9950, lastNav: 10020, mktValue: 10020000, unrealPL: 70000, plPct: 0.70, weight: 9.6 },
  { id: 'h7', security: 'AOT TB',     category: 'Equity', sector: 'Transport',    qty: 100000, avgCost: 62.0,  lastNav: 60.5,  mktValue: 6050000,  unrealPL: -150000, plPct: -2.42, weight: 5.8 },
]

const totalValue    = computed(() => holdings.reduce((s, h) => s + h.mktValue, 0))
const totalUnrealPL = computed(() => holdings.reduce((s, h) => s + h.unrealPL, 0))
const totalCost     = computed(() => holdings.reduce((s, h) => s + (h.avgCost * (h.category === 'Cash' ? 1 : h.qty)), 0))

const txns = [
  { date: '2026-03-13', type: 'SELL', security: 'PTT TB',    qty: 20000, price: 69.00, amount: 1380000, status: 'SETTLED' },
  { date: '2026-03-12', type: 'BUY',  security: 'ADVANC TB', qty: 15000, price: 240.0, amount: 3600000, status: 'SETTLED' },
  { date: '2026-03-11', type: 'BUY',  security: 'CPALL TB',  qty: 30000, price: 48.50, amount: 1455000, status: 'SETTLED' },
  { date: '2026-03-10', type: 'SELL', security: 'AOT TB',    qty: 5000,  price: 61.20, amount: 306000,  status: 'SETTLED' },
]
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <div class="breadcrumb" style="margin-bottom:4px;"><span>Portfolio</span><span class="breadcrumb-sep">/</span><span class="breadcrumb-current">Portfolio Management</span></div>
        <h1 class="page-title">Portfolio Management</h1>
        <p class="page-desc">Holdings, valuations, simulated P&amp;L, and transaction ledger — paper trading phase</p>
      </div>
      <div style="display:flex;align-items:center;gap:8px;">
        <select v-model="filterContract" class="form-select" style="height:30px;font-size:12.5px;width:140px;padding:0 8px;">
          <option>ABCFLEX1</option><option>ABCFLEX2</option><option>ABCEQ1</option><option>ABCBOND1</option>
        </select>
        <button class="btn btn-secondary btn-sm">Export</button>
      </div>
    </div>

    <!-- Stale NAV warning -->
    <div v-if="staleNAVWarn" class="alert alert-warning" style="margin-bottom:14px;">
      <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor" style="flex-shrink:0;"><path d="M8 2L1 14h14L8 2z"/></svg>
      <div><strong>Stale NAV Warning:</strong> Market data for AOT TB has not been refreshed in over 4 hours. Valuations may be inaccurate.</div>
      <button class="btn btn-warning btn-sm" style="margin-left:auto;flex-shrink:0;" @click="staleNAVWarn=false">Dismiss</button>
    </div>

    <!-- Summary Stats -->
    <div class="grid-4" style="margin-bottom:16px;">
      <div class="stat-card">
        <div class="stat-card-label">Total Portfolio Value</div>
        <div class="stat-card-value numeric" style="font-size:20px;">฿{{ (totalValue/1000000).toFixed(2) }}M</div>
        <div class="stat-card-meta">As of 13 Mar 2026 14:30</div>
      </div>
      <div class="stat-card">
        <div class="stat-card-label">Total Cost Basis</div>
        <div class="stat-card-value numeric" style="font-size:20px;">฿{{ (totalCost/1000000).toFixed(2) }}M</div>
        <div class="stat-card-meta">Average cost across holdings</div>
      </div>
      <div class="stat-card">
        <div class="stat-card-label">Unrealized P&amp;L</div>
        <div class="stat-card-value numeric" style="font-size:20px;" :class="totalUnrealPL >= 0 ? 'positive' : 'negative'">
          {{ totalUnrealPL >= 0 ? '+' : '' }}฿{{ (totalUnrealPL/1000000).toFixed(2) }}M
        </div>
        <div class="stat-card-meta" :class="totalUnrealPL >= 0 ? 'positive' : 'negative'">
          {{ ((totalUnrealPL / totalCost) * 100).toFixed(2) }}% on cost
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-card-label">Cash Position</div>
        <div class="stat-card-value numeric" style="font-size:20px;">฿12.45M</div>
        <div class="stat-card-meta">11.9% of portfolio</div>
      </div>
    </div>

    <!-- Category Allocation Mini Chart -->
    <div class="card" style="margin-bottom:16px;">
      <div class="card-header"><span class="card-title">Asset Allocation</span><span style="font-size:12px;color:#64748b;">{{ filterContract }}</span></div>
      <div class="card-body">
        <div style="display:grid;grid-template-columns:repeat(5,1fr);gap:12px;">
          <div v-for="cat in [
            { name: 'Equity',    weight: 80.7, color: '#2563eb' },
            { name: 'Bonds',     weight: 9.6,  color: '#0891b2' },
            { name: 'Cash',      weight: 11.9, color: '#64748b' },
            { name: 'Limit (Eq)', weight: 75,  color: '#dc2626', isLimit: true },
            { name: 'Limit (Bd)', weight: 30,  color: '#d97706', isLimit: true },
          ]" :key="cat.name" style="text-align:center;">
            <div style="display:flex;align-items:flex-end;justify-content:center;height:60px;margin-bottom:4px;">
              <div style="width:32px;border-radius:3px 3px 0 0;transition:height 0.3s;"
                :style="`height:${Math.min(cat.weight,100) * 0.6}px;background:${cat.isLimit ? 'repeating-linear-gradient(45deg,' + cat.color + ',' + cat.color + ' 2px,transparent 2px,transparent 6px)' : cat.color};border:${cat.isLimit ? '1px dashed ' + cat.color : 'none'};`"
              ></div>
            </div>
            <div style="font-size:13px;font-weight:700;" :style="`color:${cat.isLimit ? cat.color : '#0f172a'}`">{{ cat.weight }}%</div>
            <div style="font-size:11px;color:#94a3b8;">{{ cat.name }}</div>
          </div>
        </div>
        <div class="alert alert-error" style="margin-top:12px;">
          <svg width="13" height="13" viewBox="0 0 16 16" fill="currentColor" style="flex-shrink:0;"><path d="M8 2L1 14h14L8 2z"/></svg>
          <div style="font-size:12px;"><strong>Category limit breach:</strong> Equity allocation (80.7%) exceeds the configured IRG limit of 75%. No further BUY decisions for equity will pass IRG until rebalanced.</div>
        </div>
      </div>
    </div>

    <!-- Tabs: Holdings / Transactions -->
    <div class="card">
      <div class="card-header" style="padding-bottom:0;">
        <div class="tab-bar" style="border-bottom:none;flex:1;">
          <div class="tab-item" :class="activeTab==='holdings'?'active':''" @click="activeTab='holdings'">Holdings <span class="tab-count">{{ holdings.length }}</span></div>
          <div class="tab-item" :class="activeTab==='txns'?'active':''" @click="activeTab='txns'">Transaction Ledger <span class="tab-count">{{ txns.length }}</span></div>
        </div>
      </div>

      <!-- Holdings Table -->
      <template v-if="activeTab==='holdings'">
        <table class="data-table">
          <thead>
            <tr>
              <th class="sortable sorted">Security</th>
              <th>Category</th>
              <th>Sector</th>
              <th class="col-right">Quantity</th>
              <th class="col-right">Avg Cost</th>
              <th class="col-right">Last NAV</th>
              <th class="col-right">Mkt Value (฿)</th>
              <th class="col-right">Unrealized P&amp;L</th>
              <th class="col-right">P&amp;L %</th>
              <th class="col-right">Weight</th>
              <th class="col-action"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="h in holdings" :key="h.id"
              @click="selectedPos=h.id; showPosDrawer=true"
              style="cursor:pointer;"
              :class="h.plPct < -5 ? 'row-error' : ''"
            >
              <td style="font-weight:600;">{{ h.security }}</td>
              <td><span class="badge badge-neutral" style="font-size:10px;">{{ h.category }}</span></td>
              <td style="font-size:12.5px;color:#64748b;">{{ h.sector }}</td>
              <td class="col-right numeric">{{ h.category === 'Cash' ? '—' : h.qty.toLocaleString() }}</td>
              <td class="col-right numeric">{{ h.category === 'Cash' ? '—' : '฿' + h.avgCost.toFixed(2) }}</td>
              <td class="col-right numeric">
                <span :class="h.security === 'AOT TB' ? 'badge badge-warning' : ''" style="font-size:h.security==='AOT TB'?'11px':''">
                  ฿{{ h.lastNav.toLocaleString() }}
                </span>
              </td>
              <td class="col-right numeric" style="font-weight:600;">฿{{ h.mktValue.toLocaleString() }}</td>
              <td class="col-right numeric" :class="h.unrealPL >= 0 ? 'positive' : 'negative'">
                {{ h.unrealPL >= 0 ? '+' : '' }}฿{{ h.unrealPL.toLocaleString() }}
              </td>
              <td class="col-right numeric" :class="h.plPct >= 0 ? 'positive' : 'negative'">
                {{ h.plPct >= 0 ? '+' : '' }}{{ h.plPct.toFixed(2) }}%
              </td>
              <td class="col-right">
                <div style="display:flex;align-items:center;gap:5px;justify-content:flex-end;">
                  <div style="width:40px;height:4px;background:#f1f5f9;border-radius:2px;overflow:hidden;">
                    <div :style="`width:${h.weight}%;height:100%;background:#2563eb;`"></div>
                  </div>
                  <span class="numeric" style="font-size:12px;color:#374151;">{{ h.weight }}%</span>
                </div>
              </td>
              <td class="col-action">
                <button class="btn btn-ghost btn-icon-sm">
                  <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M1 4h14M1 8h10M1 12h6"/></svg>
                </button>
              </td>
            </tr>
          </tbody>
          <tfoot>
            <tr style="background:#f8fafc;font-weight:700;border-top:2px solid #e2e8f0;">
              <td colspan="6" style="padding:8px 12px;font-size:12.5px;color:#374151;">Total</td>
              <td class="col-right numeric" style="padding:8px 12px;">฿{{ totalValue.toLocaleString() }}</td>
              <td class="col-right numeric" style="padding:8px 12px;" :class="totalUnrealPL >= 0 ? 'positive' : 'negative'">
                {{ totalUnrealPL >= 0 ? '+' : '' }}฿{{ totalUnrealPL.toLocaleString() }}
              </td>
              <td class="col-right numeric" style="padding:8px 12px;" :class="totalUnrealPL >= 0 ? 'positive' : 'negative'">
                {{ ((totalUnrealPL / totalCost)*100).toFixed(2) }}%
              </td>
              <td style="padding:8px 12px;" class="col-right">100%</td>
              <td></td>
            </tr>
          </tfoot>
        </table>
      </template>

      <!-- Transactions Table -->
      <template v-if="activeTab==='txns'">
        <div class="card-header" style="border-top:1px solid #f1f5f9;">
          <div class="filter-bar">
            <select class="form-select" style="height:28px;font-size:12px;width:100px;padding:0 8px;"><option>All Types</option><option>BUY</option><option>SELL</option></select>
            <input class="form-input" type="date" style="height:28px;font-size:12px;width:130px;" value="2026-03-10" />
            <span style="color:#94a3b8;font-size:13px;">to</span>
            <input class="form-input" type="date" style="height:28px;font-size:12px;width:130px;" value="2026-03-13" />
          </div>
        </div>
        <table class="data-table">
          <thead><tr>
            <th>Date</th><th>Type</th><th>Security</th><th class="col-right">Quantity</th>
            <th class="col-right">Price (฿)</th><th class="col-right">Amount (฿)</th><th>Status</th>
          </tr></thead>
          <tbody>
            <tr v-for="t in txns" :key="t.date+t.security">
              <td class="col-mono" style="color:#475569;">{{ t.date }}</td>
              <td><span class="badge" :class="t.type==='BUY'?'badge-info':'badge-error'" style="font-size:10px;">{{ t.type }}</span></td>
              <td style="font-weight:600;">{{ t.security }}</td>
              <td class="col-right numeric">{{ t.qty.toLocaleString() }}</td>
              <td class="col-right numeric">฿{{ t.price.toFixed(2) }}</td>
              <td class="col-right numeric" style="font-weight:600;">฿{{ t.amount.toLocaleString() }}</td>
              <td><span class="badge badge-teal" style="font-size:10px;">{{ t.status }}</span></td>
            </tr>
          </tbody>
        </table>
      </template>
      <div class="card-footer" style="display:flex;justify-content:space-between;align-items:center;">
        <span style="font-size:12px;color:#64748b;">{{ activeTab==='holdings' ? holdings.length + ' positions' : txns.length + ' transactions' }}</span>
        <button class="btn btn-secondary btn-sm">Export CSV</button>
      </div>
    </div>

    <!-- Position Detail Drawer -->
    <template v-if="showPosDrawer && selectedPos">
      <div class="drawer-overlay" @click="showPosDrawer=false"></div>
      <div class="drawer">
        <div class="drawer-header">
          <div>
            <div class="drawer-title">Position Detail</div>
            <div class="drawer-subtitle">{{ holdings.find(h=>h.id===selectedPos)?.security }} &nbsp;·&nbsp; {{ filterContract }}</div>
          </div>
          <button class="btn btn-ghost btn-icon-sm" @click="showPosDrawer=false"><svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 3l10 10M13 3L3 13"/></svg></button>
        </div>
        <div class="drawer-body" style="display:flex;flex-direction:column;gap:16px;" v-if="holdings.find(h=>h.id===selectedPos) as any">
          <div style="display:grid;grid-template-columns:1fr 1fr;gap:10px;">
            <div style="padding:10px 12px;background:#f8fafc;border-radius:5px;border:1px solid #e2e8f0;" v-for="field in [
              {label:'Quantity', value: holdings.find(h=>h.id===selectedPos)?.qty?.toLocaleString()},
              {label:'Avg Cost', value:'฿' + holdings.find(h=>h.id===selectedPos)?.avgCost?.toFixed(2)},
              {label:'Last NAV', value:'฿' + holdings.find(h=>h.id===selectedPos)?.lastNav?.toLocaleString()},
              {label:'Market Value', value:'฿' + holdings.find(h=>h.id===selectedPos)?.mktValue?.toLocaleString()},
              {label:'Unrealized P&L', value:(holdings.find(h=>h.id===selectedPos)?.unrealPL??0)>=0?'+฿'+(holdings.find(h=>h.id===selectedPos)?.unrealPL??0).toLocaleString():'฿'+(holdings.find(h=>h.id===selectedPos)?.unrealPL??0).toLocaleString()},
              {label:'P&L %', value:((holdings.find(h=>h.id===selectedPos)?.plPct??0)>=0?'+':'')+holdings.find(h=>h.id===selectedPos)?.plPct?.toFixed(2)+'%'},
            ]" :key="field.label">
              <div class="text-label" style="margin-bottom:3px;">{{ field.label }}</div>
              <div class="numeric" style="font-size:14px;font-weight:600;">{{ field.value }}</div>
            </div>
          </div>
          <div>
            <div class="text-label" style="margin-bottom:6px;">Valuation History (7 days)</div>
            <div style="display:flex;align-items:flex-end;gap:3px;height:60px;padding:0 4px;">
              <div v-for="(v,i) in [44.8,45.2,46.1,47.5,48.0,48.8,49.1]" :key="i"
                style="flex:1;background:#dbeafe;border-radius:2px 2px 0 0;min-height:4px;"
                :style="`height:${((v-44)/6)*100}%;`"
                :title="'฿' + v"
              ></div>
            </div>
            <div style="display:flex;justify-content:space-between;font-size:10px;color:#94a3b8;margin-top:3px;">
              <span>Mar 7</span><span>Mar 8</span><span>Mar 9</span><span>Mar 10</span><span>Mar 11</span><span>Mar 12</span><span>Mar 13</span>
            </div>
          </div>
        </div>
        <div class="drawer-footer">
          <button class="btn btn-secondary btn-sm" @click="showPosDrawer=false">Close</button>
          <button class="btn btn-primary btn-sm">Create Decision</button>
        </div>
      </div>
    </template>
  </div>
</template>
