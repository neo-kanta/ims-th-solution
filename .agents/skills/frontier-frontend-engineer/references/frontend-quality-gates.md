# Frontend Quality Gates

Load this reference for implementation planning, UI review, or final frontend
verification.

## Workflow and State

- Identify user role, portfolio type, route, business date, and primary action.
- Verify loading, empty, error, permission, disabled, stale, success, and retry
  states that can occur.
- Prevent duplicate submissions and preserve server idempotency signals.
- Confirm destructive or irreversible actions require clear user intent.
- Keep backend rejection details useful without leaking internals.

## API and Identity

- Generate and consume `ims-api.d.ts`; never duplicate DTOs manually.
- Keep request construction inside typed feature services.
- Do not infer permissions or financial policy from hidden buttons.
- Prefer portfolio code and business labels; do not display internal UUIDs.
- Surface a backend contract gap instead of inventing mock production behavior.

## Localization and Content

- Add EN/TH/ZH keys together and verify each changed path.
- Keep financial terminology consistent with repository domain docs.
- Test long Thai and Chinese text in compact controls and tables.
- Do not expose implementation terminology such as raw scope or ID field names.

## Accessibility and Responsive Behavior

- Use semantic controls, keyboard navigation, visible focus, associated labels,
  and announced validation/errors.
- Check color contrast and do not communicate state by color alone.
- Respect reduced motion and avoid layout-shifting dynamic content.
- Verify narrow mobile, standard desktop, and wide desktop layouts.
- Check table overflow, modals, menus, sticky elements, and text wrapping.

## Browser Proof

- Start the correct local app against an authorized backend/test environment.
- Exercise the real route and primary workflow, not only a component preview.
- Inspect console errors, failed network requests, and response/error states.
- Capture screenshots at desktop and mobile for material visual changes.
- Verify feature behavior with permissions and representative synthetic data.

## Minimum Evidence by Risk

| Risk | Minimum evidence |
| --- | --- |
| Copy or isolated styling | Diff review, affected locale/render check |
| Component or service behavior | Focused Vitest, full build, type diagnostic comparison |
| Financial workflow or shared shell | Full frontend tests/build, browser proof, independent review, backend-contract confirmation |

## Completion Report

```text
User-visible outcome:
Routes/components/services:
Generated API client status:
States and locales verified:
Commands run and results:
Browser viewports and evidence:
Not proved or deferred:
Git state:
Backend/integration request:
```
