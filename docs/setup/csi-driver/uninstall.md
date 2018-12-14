---
title: Uninstall
description: Vault CSI Driver Uninstall
menu:
  product_vault:
    identifier: uninstall-csi-driver
    name: Uninstall
    parent: setup
    weight: 10
product_name: csi-driver
menu_name: product_vault
section_menu_id: setup
---

# Uninstall Vault CSI Driver

If you installed csi driver using YAML then run:

```console
$ curl -fsSL https://raw.githubusercontent.com/kubevault/csi-driver/0.1.0/hack/deploy/install.sh \
    | bash -s -- --uninstall [--namespace=NAMESPACE]

```

The above command will leave the csidriver crd objects as-is. If you wish to nuke all csidriver crd objects, also pass the `--purge` flag.

If you used HELM to install Vault CSI driver, then run following command

```console
helm del --purge <name>
```
