# IMS Thailand - Frontend Component Inventory

This document lists all shared Application Components available in the frontend codebase. These components are registered globally (via Nuxt auto-imports under `app/shared/ui/` with `pathPrefix: false`) and are ready to be used in any feature screen.

---

## Base UI Components

### 1. AppPage (`AppPage.vue`)
* **Purpose**: Reusable wrapper for standard layout screens. Handles page structure, title, subtitle, breadcrumbs, loading state, error display, and sidebar rails.
* **Props**:
  * `title?: string` (Page header title)
  * `subtitle?: string` (Page description)
  * `loading?: boolean` (Toggle spinner loading state for the entire page body)
  * `error?: string | null` (Display error banner with retry trigger)
  * `width?: "full" | "contained" | "narrow"` (Contained is `max-width: 1280px`, narrow is `max-width: 800px`, full is `100%`)
* **Slots**:
  * `eyebrow` (Eyebrow/breadcrumb tag at the top left)
  * `actions` (Actions section on the right side of the header)
  * `tabs` (Tab panel directly under the page title)
  * `default` (Main screen content)
  * `right-rail` (Optional sidebar slot)
* **Usage**:
  ```vue
  <template>
    <AppPage title="Accounts" :loading="loading" :error="error" @retry="fetchData">
      <template #eyebrow>
        <span>Administration</span>
      </template>
      <template #actions>
        <AppButton variant="primary">Create User</AppButton>
      </template>
      <div class="card">Main page body content</div>
    </AppPage>
  </template>
  ```

### 2. AppSection (`AppSection.vue`)
* **Purpose**: Structural section groups within pages (e.g. settings panels, portfolio subsections).
* **Props**:
  * `title?: string`
  * `description?: string`
  * `variant?: "bordered" | "plain" | "elevated"` (Default is `"bordered"`)
* **Slots**:
  * `default`
  * `actions` (Actions nested on the section header line)

### 3. AppPanel (`AppPanel.vue`)
* **Purpose**: General pane interface, complete with header/footer containers, supporting multiple densities, selection borders, and disabled styles.
* **Props**:
  * `title?: string`
  * `subtitle?: string`
  * `density?: "compact" | "comfortable"` (Default is `"comfortable"`)
  * `selected?: boolean`
  * `disabled?: boolean`
* **Slots**:
  * `header-actions`
  * `default`
  * `footer`

### 4. AppButton (`AppButton.vue`)
* **Purpose**: Standard button control conforming to the theme palette.
* **Props**:
  * `variant?: "primary" | "secondary" | "danger" | "warning" | "success" | "ghost"` (Default is `"secondary"`)
  * `size?: "xs" | "sm" | "md" | "lg"` (Default is `"md"`)
  * `loading?: boolean`
  * `disabled?: boolean`
  * `icon?: boolean` (Reduces padding to make it a square icon wrapper)
  * `type?: "button" | "submit" | "reset"` (Default is `"button"`)
  * `fullWidth?: boolean` (Expands width to 100%)

### 5. AppActionBar (`AppActionBar.vue`)
* **Purpose**: Bottom layout footer for action rows (forms save/cancel, approvals submission, delete confirmations). Can be made sticky.
* **Props**:
  * `actions?: Array<"save" | "cancel" | "submit" | "review" | "cancel-review" | "delete">`
  * `saveLabel?: string`, `cancelLabel?: string`, etc. (Custom overrides)
  * `loadingAction?: string | null` (Specific action type that shows a loading spinner)
  * `disabled?: boolean`
  * `sticky?: boolean`

### 6. AppIconButton (`AppIconButton.vue`)
* **Purpose**: Compact square button presenting an icon. Enforces accessibility via a required `ariaLabel` prop.
* **Props**:
  * `icon: string` (Icon SVG glyph name matching `AppIcon`)
  * `ariaLabel: string` (Enforced accessibility label)
  * `tooltip?: string` (Tooltip helper hint)
  * `variant?`, `size?`, `disabled?`, `loading?` (Matches `AppButton`)

### 7. AppToastProvider & `useAppToast`
* **Purpose**: Global context-wide toast notification queue. Avoids template boilerplate.
* **API Details (`useAppToast`)**:
  * `showSuccess(message: string, duration?: number)`
  * `showError(message: string, duration?: number)`
  * `showWarning(message: string, duration?: number)`
  * `showInfo(message: string, duration?: number)`
  * `dismiss(id: number)`
  * `clearAll()`
* **Usage**:
  ```ts
  const toast = useAppToast();
  toast.showSuccess("Settings updated successfully!");
  ```

### 8. AppDataTable (`AppDataTable.vue`)
* **Purpose**: Standard enterprise grid. Supporting sortable headers, single/multi selection, customized templates, and empty/loading states.
* **Props**:
  * `columns: Array<{ key: string, label: string, sortable?: boolean, align?: "left" | "center" | "right", width?: string }>`
  * `items: any[]`
  * `loading?: boolean`
  * `error?: string | null`
  * `sortBy?: string`, `sortOrder?: "asc" | "desc"`
  * `rowClickable?: boolean`
  * `selectedIds?: any[]`
* **Slots**:
  * `cell(colKey)`: Scoped slot to custom render a cell. Exposes `{ item, value }`.
  * `pagination`: Bottom container for pagination footer selectors.
* **Usage**:
  ```vue
  <template>
    <AppDataTable :columns="cols" :items="records">
      <template #cell(status)="{ item }">
        <AppStatusBadge :status="item.status" />
      </template>
    </AppDataTable>
  </template>
  ```

---

## Form Primitives

All form inputs integrate with Vue's `v-model` and validation boundaries (red borders/error indicators).

* **`AppFormField`**: Input decorator supporting required marks, labels, help hints, and alignment layouts.
* **`AppInput`**: Wraps `<input>` with support for text, numbers, search, passwords, and emails.
* **`AppSelect`**: Wraps `<select>` dropdowns supporting array options.
* **`AppDateField`**: Native calendar input.
* **`AppDateRangeField`**: Double date picker inputs using multiple models (`v-model:start` and `v-model:end`).
* **`AppTextarea`**: Multiline input.
* **`AppCheckbox`**: Standard tick selector.
* **`AppRadioGroup`**: Grouped radio button list.
* **`AppSearchPanel`**: Multi-input filter cards providing Search/Apply and Clear action columns.

---

## FinTech / Formatting Components

* **`AppMoney`**: Formats currencies (e.g. THB, USD) with decimals and support for negative accounting parenthesizing.
* **`AppPercent`**: Formats signed percentages with red/green colors for up/down changes.

---

## IMS Domain Components

* **`IMSWorkflowStageTracker`**: 4-stage visual pipeline tracker (Day Start -> Approval -> Tx Close -> Acct Close) with metadata/stamps.
* **`IMSApprovalPanel`**: Approvals audit logger displaying signs, delegators, and reviewer tags.
* **`IMSApprovalStamp`**: Digital sign-off seal displaying approver name, timestamp, and delegation markers `(代)`.
* **`IMSDelegationBanner`**: Header notices warning operators that they are currently acting on behalf of a principal account.
* **`IMSPermissionGuard`**: Client authorization guards checking if user possesses the required function permission codes.
* **`IMSAuditTimeline`**: Compact lists displaying actor actions and nested details.
* **`IMSComplianceResultPanel`**: Results lists detailing pre/post validation thresholds and breaches.
* **`IMSContractSelector`**: Searchable selection panels for active contracts.
* **`IMSOperationDatePanel`**: Details panel showing the active business operating date.

---

## Dos and Don'ts

### Dos:
- **Do** use `AppPage` for all main feature screens. It guarantees consistent loading and error handling.
- **Do** always provide an `ariaLabel` to `AppIconButton` to prevent screen reader warnings.
- **Do** place currency calculations in stores and use `AppMoney` or `AppPercent` solely for display styling.
- **Do** wrap destructive actions (like Delete) in an `AppConfirmDialog` with a warning tone.

### Don'ts:
- **Don't** hardcode currency characters (like `฿` or `$`) in templates; let `AppMoney` format it based on the currency prop.
- **Don't** construct tables manually using raw `<table>` blocks unless `AppDataTable` is structurally incompatible.
- **Don't** declare local refs for toasts inside features if you can use `useAppToast()` and `AppToastProvider`.
