---
name: Post Test — Coverage Report
trigger: post_test
---

### Actions

1. **Parse coverage output** — If `coverage.out` exists, calculate per-package coverage
2. **Report coverage summary:**

```
📊 Coverage Report
──────────────────
internal/domain     95.2%  ✅
internal/service    87.3%  ✅
internal/handler    82.1%  ⚠️ (below 85%)
internal/repository 78.4%  ❌ (below 85%)
internal/config     100%   ✅
──────────────────
TOTAL               86.7%  ✅
```

1. **Identify uncovered functions** — List any public functions with 0% coverage
2. **Track coverage trend** — Warn if coverage decreased from last run