package yamls

const CONFIG_ENV = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: account-iam-env-configmap-dev
  namespace: mcsp
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
  annotations:
    argocd.argoproj.io/sync-wave: "0"
data:
  NOTIFICATION_SERVICE_ENABLED: ""
  LOCAL_TOKEN_ISSUER: https://mcspid/account-iam/api/2.0
`

const CONFIG_JWT = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: account-iam
  namespace: mcsp
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
  annotations:
    argocd.argoproj.io/sync-wave: "0"
data:

  jwt.suffix.issuer: https://mcspid/account-iam/api/2.0
`
