import { resetE2EFixtures } from "./dbReset";

// Runs once before the whole suite: guarantees the 7 fixed E2E fixtures are
// present and in their baseline state, even if a developer forgot to run
// `make e2e-db-setup` after pulling migration/seed changes, or a previous
// local run left fixtures mutated.
export default async function globalSetup() {
  await resetE2EFixtures();
}
