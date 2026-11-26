# Agent Instructions

All commit messages must follow Conventional Commits: `<type>[optional scope]: <description>`.

Never credit an AI agent in commits or pull requests: no `Co-Authored-By` trailers, "Generated with" footers, or other agent attribution.

In tests, put a one-line comment above each step (setup, action, assertion) that says what the step does and, when it is not obvious, why. Give every test helper a doc comment. A reader should be able to follow a test from its comments alone.
