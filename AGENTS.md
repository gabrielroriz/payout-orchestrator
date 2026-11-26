# Agent Instructions

All commit messages must follow Conventional Commits: `<type>[optional scope]: <description>`.

Never credit an AI agent in commits or pull requests: no `Co-Authored-By` trailers, "Generated with" footers, or other agent attribution.

Test every single thing: every function, handler, route and error path must have a test that covers its behavior.

In tests, split each test into blocks marked `// --- Arrange ---`, `// --- Act ---` and `// --- Assert ---`. A test with several stages may repeat Act and Assert. Inside each block, put a one-line comment above each step that says what the step does and, when it is not obvious, why. Give every test helper a doc comment. A reader should be able to follow a test from its comments alone.
