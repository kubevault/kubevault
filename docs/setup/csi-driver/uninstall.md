---
title: Uninstall Vault CSI Driver
menu:
  docs_0.1.0:
    identifier: uninstall-csi-driver
    name: Uninstall
    parent: csi-driver-setup
    weight: 15
menu_name: docs_0.1.0
section_menu_id: setup
---

# Uninstall Vault CSI Driver

If you installed csi driver using YAML then run:

```console
$ curl -fsSL https://raw.githubusercontent.com/kubevault/csi-driver/0.2.0/hack/deploy/install.sh \
    | bash -s -- --uninstall [--namespace=NAMESPACE]

```

> N.B: For Kubernetes v1.12 use `0.1.0`

The above command will leave the csidriver crd objects as-is. If you wish to nuke all csidriver crd objects, also pass the `--purge` flag.

If you used HELM to install Vault CSI driver, then run following command

```console
helm del --purge <name>
```
