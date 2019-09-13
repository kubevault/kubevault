---
title: Vault-Operator
menu:
  docs_{{ .version }}:
    identifier: vault-operator
    name: Vault-Operator
    parent: reference-operator
    weight: 0

menu_name: docs_{{ .version }}
section_menu_id: reference
url: /docs/{{ .version }}/reference/operator/
aliases:
- /docs/{{ .version }}/reference/operator/vault-operator/
---
## vault-operator

Vault Operator by AppsCode - HashiCorp Vault Operator for Kubernetes

### Synopsis

Vault Operator by AppsCode - HashiCorp Vault Operator for Kubernetes

### Options

```
      --alsologtostderr                  log to standard error as well as files
      --bypass-validating-webhook-xray   if true, bypasses validating webhook xray checks
      --enable-analytics                 Send analytical events to Google Analytics (default true)
  -h, --help                             help for vault-operator
      --log-flush-frequency duration     Maximum number of seconds between log flushes (default 5s)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --stderrthreshold severity         logs at or above this threshold go to stderr
      --test.bench regexp                run only benchmarks matching regexp
      --test.benchmem                    print memory allocations for benchmarks
      --test.benchtime d                 run each benchmark for duration d (default 1s)
      --test.blockprofile file           write a goroutine blocking profile to file
      --test.blockprofilerate rate       set blocking profile rate (see runtime.SetBlockProfileRate) (default 1)
      --test.count n                     run tests and benchmarks n times (default 1)
      --test.coverprofile file           write a coverage profile to file
      --test.cpu list                    comma-separated list of cpu counts to run each test with
      --test.cpuprofile file             write a cpu profile to file
      --test.failfast                    do not start new tests after the first test failure
      --test.list regexp                 list tests, examples, and benchmarks matching regexp then exit
      --test.memprofile file             write an allocation profile to file
      --test.memprofilerate rate         set memory allocation profiling rate (see runtime.MemProfileRate)
      --test.mutexprofile string         write a mutex contention profile to the named file after execution
      --test.mutexprofilefraction int    if >= 0, calls runtime.SetMutexProfileFraction() (default 1)
      --test.outputdir dir               write profiles to dir
      --test.parallel n                  run at most n tests in parallel (default 8)
      --test.run regexp                  run only tests and examples matching regexp
      --test.short                       run smaller test suite to save time
      --test.testlogfile file            write test action log to file (for use only by cmd/go)
      --test.timeout d                   panic test binary after duration d (default 0, timeout disabled) (default 0s)
      --test.trace file                  write an execution trace to file
      --test.v                           verbose: print additional output
      --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [vault-operator run](/docs/reference/operator/vault-operator_run.md)	 - Launch Vault operator
* [vault-operator version](/docs/reference/operator/vault-operator_version.md)	 - Prints binary version number.

