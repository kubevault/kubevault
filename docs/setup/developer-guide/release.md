---
title: Release | Vault operator
menu:
  docs_0.2.0:
    identifier: release
    name: Release
    parent: developer-guide
    weight: 15
menu_name: docs_0.2.0
section_menu_id: setup
---
# Release Process

The following steps must be done from a Linux x64 bit machine.

- Do a global replacement of tags so that docs point to the next release.
- Push changes to the `release-x` branch and apply new tag.
- Push all the changes to remote repo.
- Build and push vault docker image:
```console
$ cd ~/go/src/github.com/appscode/vault
./hack/docker/setup.sh; env APPSCODE_ENV=prod ./hack/docker/setup.sh release
```

- Now, update the release notes in Github. See previous release notes to get an idea what to include there.
