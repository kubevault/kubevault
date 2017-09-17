#!/usr/bin/env python


# http://stackoverflow.com/a/14050282
def check_antipackage():
    from sys import version_info
    sys_version = version_info[:2]
    found = True
    if sys_version < (3, 0):
        # 'python 2'
        from pkgutil import find_loader
        found = find_loader('antipackage') is not None
    elif sys_version <= (3, 3):
        # 'python <= 3.3'
        from importlib import find_loader
        found = find_loader('antipackage') is not None
    else:
        # 'python >= 3.4'
        from importlib import util
        found = util.find_spec('antipackage') is not None
    if not found:
        print('Install missing package "antipackage"')
        print('Example: pip install git+https://github.com/ellisonbg/antipackage.git#egg=antipackage')
        from sys import exit
        exit(1)
check_antipackage()

# ref: https://github.com/ellisonbg/antipackage
import antipackage
from github.appscode.libbuild import libbuild, pydotenv

import os
import os.path
import subprocess
import sys
import time
import yaml
from os.path import expandvars, join, dirname

libbuild.REPO_ROOT = expandvars('$GOPATH') + '/src/github.com/appscode/steward'
BUILD_METADATA = libbuild.metadata(libbuild.REPO_ROOT)
libbuild.BIN_MATRIX = {
    'steward': {
        'type': 'go',
        'go_version': True,
        'release': True,
        'distro': {
            'alpine': ['amd64'],
            'linux': ['amd64']
        }
    }
}
if libbuild.ENV not in ['prod']:
    libbuild.BIN_MATRIX['steward']['distro'] = {
        'alpine': ['amd64']
    }
libbuild.BUCKET_MATRIX = {
    'prod': 'gs://appscode-cdn',
    'dev': 'gs://appscode-dev'
}


def call(cmd, stdin=None, cwd=libbuild.REPO_ROOT):
    print(cmd)
    return subprocess.call([expandvars(cmd)], shell=True, stdin=stdin, cwd=cwd)


def die(status):
    if status:
        sys.exit(status)


def check_output(cmd, stdin=None, cwd=libbuild.REPO_ROOT):
    print(cmd)
    return subprocess.check_output([expandvars(cmd)], shell=True, stdin=stdin, cwd=cwd)


def version():
    # json.dump(BUILD_METADATA, sys.stdout, sort_keys=True, indent=2)
    for k in sorted(BUILD_METADATA):
        print(k + '=' + BUILD_METADATA[k])


def fmt():
    libbuild.ungroup_go_imports('*.go', 'pkg')
    die(call('goimports -w *.go pkg'))
    call('gofmt -s -w *.go pkg')


def vet():
    call('go vet *.go ./...')


def lint():
    call('golint *.go ./...')


def gen():
    call('go vet ./...')


def build_cmd(name):
    cfg = libbuild.BIN_MATRIX[name]
    if cfg['type'] == 'go':
        if 'distro' in cfg:
            for goos, archs in cfg['distro'].items():
                for goarch in archs:
                    libbuild.go_build(name, goos, goarch, main='*.go')
        else:
            libbuild.go_build(name, libbuild.GOHOSTOS, libbuild.GOHOSTARCH, main='*.go')


def build_cmds():
    gen()
    for name in libbuild.BIN_MATRIX:
        build_cmd(name)


def build(name=None):
    if name:
        cfg = libbuild.BIN_MATRIX[name]
        if cfg['type'] == 'go':
            gen()
            build_cmd(name)
    else:
        build_cmds()


def push_bin(bindir):
    call('rm -f *.md5', cwd=bindir)
    call('rm -f *.sha1', cwd=bindir)
    for f in os.listdir(bindir):
        if os.path.isfile(bindir + '/' + f):
            libbuild.upload_to_cloud(bindir, f, BUILD_METADATA['version'])


def push(name=None):
    if name:
        bindir = libbuild.REPO_ROOT + '/dist/' + name
        push_bin(bindir)
    else:
        dist = libbuild.REPO_ROOT + '/dist'
        for name in os.listdir(dist):
            d = dist + '/' + name
            if os.path.isdir(d):
                push_bin(d)


def update_registry():
    vf = libbuild.REPO_ROOT + '/dist/steward/versions.json'
    bucket = libbuild.BUCKET_MATRIX.get(libbuild.ENV, libbuild.BUCKET_MATRIX['dev'])
    call('gsutil cp {0}/binaries/steward/versions.json {1}'.format(bucket, vf))
    vj = {}
    if os.path.isfile(vf):
        vj = libbuild.read_json(vf)
    vj[BUILD_METADATA['version']] = {
        'changesets': [],
        'release_date': int(time.time())
    }
    libbuild.write_json(vj, vf)
    call("gsutil cp {1} {0}/binaries/steward/versions.json".format(bucket, vf))
    call('gsutil acl ch -u AllUsers:R -r {0}/binaries/steward/versions.json'.format(bucket))

    lf = libbuild.REPO_ROOT + '/dist/steward/latest.txt'
    libbuild.write_file(lf, BUILD_METADATA['version'])
    call("gsutil cp {1} {0}/binaries/steward/latest.txt".format(bucket, lf))
    call('gsutil acl ch -u AllUsers:R -r {0}/binaries/steward/latest.txt'.format(bucket))


def install():
    die(call('GO15VENDOREXPERIMENT=1 ' + libbuild.GOC + ' install ./...'))


def default():
    gen()
    fmt()
    die(call('GO15VENDOREXPERIMENT=1 ' + libbuild.GOC + ' install .'))


if __name__ == "__main__":
    if len(sys.argv) > 1:
        # http://stackoverflow.com/a/834451
        # http://stackoverflow.com/a/817296
        globals()[sys.argv[1]](*sys.argv[2:])
    else:
        default()
