package yamls

const INGRESS = `kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: account-iam-ingress-allow
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
  annotations:
    argocd.argoproj.io/sync-wave: "3"
spec:
  podSelector:
    matchLabels:
      name: account-iam 
  ingress:
    - ports:
        # calls to the API
        - protocol: TCP
          port: 9445 
  policyTypes:
    - Ingress
`

const EGRESS = `
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: account-iam-egress-allow
  labels:
    bcdr-candidate: t
    component-name: iam-services
    by-squad: mcsp-user-management
    for-product: all
  annotations:
    argocd.argoproj.io/sync-wave: "0"
spec:
  podSelector:
    matchLabels:
      name: account-iam 
  policyTypes:
    - Egress
  egress:
    # calls to openshift's dns
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: openshift-dns
    # other calls
    - ports:
        # default https calls
        - protocol: TCP
          port: 443
        # okd external route - temporary
        - protocol: TCP
          port: 6443
        # Instana agent
        - protocol: TCP
          port: 42699
        # Ephemeral port range for Instana JVM agent
        - protocol: TCP
          port: 32768
          endPort: 65535
        # Port for PG DB
        - protocol: TCP
          port: 5432`
