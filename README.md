# boogo

> [!WARNING]
> ⚠️ This is an experimental research project.

---

## What is Boogo?

Boogie is an intermediate verification language used by tools such as **Dafny**. While Boogie is excellent for verification, it is **not executable**.

**Boogo makes verified Boogie programs runnable** by compiling an executable subset of Boogie directly into native Go code.

```bash
  Dafny
    ↓
  Boogie (verified)
    ↓ [Executable Subset]
  Boogo
    ↓
  Go (executable)
```

Boogo does **not** interpret Boogie and does **not** use solvers at runtime.  
The output is plain Go code compiled by the Go toolchain.

Ok? but why do we need executable Boogie?

- Boogie is the verification boundary: it exposes explicit control flow and heap semantics after verification, making it the right level to execute verified programs.

- Verification and execution are different problems: an executable Boogie subset bridges the gap, avoiding reimplementation and mismatches between verified and deployed code.

- Not all Boogie is executable: defining an explicit executable subset cleanly separates solver-only constructs from real runtime semantics.

## Design Goals

- Compile verified programs into **real executable code**
- Preserve **partial correctness** guarantees
- Ensure **deterministic execution**
- Erase all **verification-only constructs**
- Keep the supported language **small and explicit**

## Executable Boogie Subset (EBS)

Boogo supports a restricted subset of Boogie with a clear executable meaning.

### Supported

- `procedure` and `call`
- Local variables and assignment
- `int`, `bool`
- Structured control flow (`if`, `while`)
- Deterministic expressions
- Restricted heap read/write
- Assertions (erased before execution)

### Not Supported

- `havoc` / nondeterminism
- Quantifiers
- Ghost state at runtime
- Uninterpreted functions
- Concurrency
- Recursion

## Why This Works

Every supported Boogie construct maps directly to Go:

| Boogie      | Go             |
| ----------- | -------------- |
| `procedure` | `func`         |
| `if/while`  | `if/for`       |
| Heap access | `map` / struct |
| `int/bool`  | `int/bool`     |

There is no symbolic execution, no solver, and no runtime verification.

## Project Status

- ✔ Executable subset defined
- ✔ Static rejection of unsupported features
- 🚧 Control-flow restructuring
- 🚧 Heap encoding
- 🚧 Code generation

## Scope & Limitations

Boogo intentionally targets a **small executable core** of Boogie.  
Many verification-oriented features do not have a direct executable interpretation and are excluded.

Extending support for richer Boogie features is future work.

## Intended Use

- Research and experimentation
- Running verified algorithms
- Studying verification-to-execution pipelines
- Case studies for PL research

## License

MIT (or specify)

## Citation

> _An Executable Subset of the Boogie Intermediate Language_  
> (Work in progress)
