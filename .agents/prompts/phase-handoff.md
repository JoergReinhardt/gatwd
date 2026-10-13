# Prompt: write a phase handoff for a fresh context

**Task:** the work on `<project>` is stopping after `<phase/step>`; write `docs/phase<N>_handoff.md`.

**Required sections, in order**
1. **Header** — one paragraph: what this is, which earlier handoff it **supersedes**, which docs to
   read first (design doc sections).
2. **Where we are (done, committed)** — capability list, not a diff; name the tests that prove it.
3. **Goals for the next phase** — numbered, each with the *open design question* and the options in
   order of preference ("pick one with the user before building").
4. **Authoring shape** — copy-pasteable patterns for the kind of code the next phase writes.
5. **Gotchas (don't re-derive)** — every trap hit so far, with the working workaround.
6. **Constraints (hard rules)** — the relevant mandates, forbidden directories/imports.
7. **Verification gates** — the exact commands.
8. **Pointers** — files, design-doc sections, memory notes.
9. **First steps** — ordered, concrete.

**Rules**
- Refer to sections, test names and file paths — **not** line numbers, `HEAD` hashes or
  assertion counts; they rot.
- Mark the superseded handoff (banner) or delete it in the commit that finishes its phase.
- Update the design doc's phase entry and Appendix B checklist in the same change.
- Say plainly what is uncommitted and what is unpushed.

**Definition of done:** a reader with no chat history can start step 1 without asking anything.
