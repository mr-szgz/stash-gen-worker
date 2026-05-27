# Grounding context from the prior conversation

## Why this repo exists

The user wants a standalone worker executable that can generate scene assets exactly like Stash routines, including Windows worker-node deployment, without running the full Stash server.

## Agreed direction

- Keep the worker in its own repository.
- Reuse official Stash routines as much as practical.
- Vendor only the minimal upstream files needed.
- Make upstream updates easy via copy/overwrite sync scripts and a manual GitHub Actions workflow that opens a PR.

## Key implementation constraints

- Avoid a full Stash fork unless the dependency cone becomes too large.
- Avoid `replace ../stash` or any build strategy that requires a sibling checkout.
- Keep documentation practical so future agents can quickly understand the repo and operate safely.

## Current local realities

- The original baseline failed because the repo required `../stash`.
- The worker now aims to compile against local vendored packages under `third_party/stash/`.
- Screenshot output handling was previously rough; worker-owned path helpers should be preferred for user-facing output reporting.
