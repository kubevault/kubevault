# Release Process

The following steps must be done from a Linux x64 bit machine.

- Do a global replacement of tags so that docs point to the next release.
- Push changes to the `release-x` branch and apply new tag.
- Push all the changes to remote repo.
- Build and push steward docker image:
```console
$ cd ~/go/src/github.com/soter/vault-operator
./hack/release.sh
```

- Now, update the release notes in Github. See previous release notes to get an idea what to include there.
