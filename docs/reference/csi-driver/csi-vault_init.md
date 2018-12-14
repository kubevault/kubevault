---
title: Csi-Vault Init
menu:
  product_kubevault_0.1.0:
    identifier: csi-vault-init
    name: Csi-Vault Init
    parent: reference-csi-driver
product_name: kubevault
menu_name: product_kubevault_0.1.0
section_menu_id: reference
---
## csi-vault init

Initializes the driver.

### Synopsis

Initializes the driver.

```
csi-vault init [flags]
```

### Options

```
      --endpoint string   CSI endpoint (default "unix:///var/lib/kubelet/plugins/com.kubevault.csi.secrets/csi.sock")
  -h, --help              help for init
      --node string       Linode Hostname (default "beast")
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --enable-analytics                 Send analytical events to Google Analytics (default true)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files (default true)
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
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [csi-vault](/docs/reference/csi-driver/csi-vault.md)	 - Vault CSI by Appscode - Start farms

