apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: ip-pool
spec:
  addresses:
  - {{ (ds "data").startIP }}-{{ (ds "data").endIP }}
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: empty