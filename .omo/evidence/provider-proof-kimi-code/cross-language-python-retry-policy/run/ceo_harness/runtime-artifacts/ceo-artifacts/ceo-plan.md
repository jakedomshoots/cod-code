# CEO Execution Plan

Authority: ceo
Mode: delegated
Next action: answer subagent questions

1. coder - apply bounded changes [needs_input]
   Applying patch and creating evidence file
2. coder - fix failed verification with minimal patch [pass]
   Updated scripts/retry_policy.py benchmark_fixture to include required terms (exponential backoff, jitter, timeout) and created the required evidence markdown.
3. checker - run verification checks [pass]
   2 check attempt(s)
4. ceo - final verdict [needs_input]
   CEO final verdict for "Complete benchmark task cross-language-python-retry-policy: Repair Python retry policy fixture so timeout retries include exponential backoff and jitter evidence..\nEdit only the required files unless an evidence artifact is required.\nRequired changed files: scripts/retry_policy.py.\nRequired evidence artifacts: .omo/evidence/cross-language-python-retry-policy.md.\nCreate every required evidence artifact as a non-empty markdown file inside the workspace.\nEach evidence artifact must summarize the change, commands run, and verification result.\nRequired diff terms: exponential backoff, jitter, timeout.\nRequired commands to satisfy after the edit: python3 scripts/test_retry_policy.py.\nDo not inspect unrelated files or run broad test suites; run only the required commands for verification.\nStop as soon as the required files, evidence artifacts, diff terms, and commands are satisfied.\nKeep the change minimal and do not remove the Go fixture files."
