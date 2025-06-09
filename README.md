# Tartarus

## Github actions

### Jobs

On pull request verifies commit messages, runs tests.
During push to master have additional step to build docker application and push it to GHCR.

### Merge pull request

Merge pull request using rebase. Otherwise it will create commit which starts "Merge*" and version bump for docker image won't work properly.

## Commits

### Commit format

*commit_type*(task):*commit message*

### Allowed commit types

| Commit type | Version increment |
| --- | --- |
| epic | major |
| major | major |
| feat | minor |
| build | patch |
| chore | patch |
| cicd | patch |
| fix | patch |
| perf | patch |
| refactor | patch |
| revert | patch |
| style | patch |
| docs | none |
| test | none |

