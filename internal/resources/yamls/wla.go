package yamls

const ACCOUNT_IAM_APP = `
apiVersion: liberty.websphere.ibm.com/v1
kind: WebSphereLibertyApplication
metadata:
  name: account-iam
  namespace: mcsp
  labels:
    by-squad: mcsp-user-management
    for-product: all
    name: account-iam
    bcdr-candidate: t
    component-name: iam-services
  annotations:
    argocd.argoproj.io/sync-wave: "3"
spec:
  license:
    accept: true
    edition: IBM WebSphere Application Server
    productEntitlementSource: Standalone
  securityContext:
    privileged: false
    runAsNonRoot: true
    readOnlyRootFilesystem: false
    allowPrivilegeEscalation: false
    seccompProfile:
      type: RuntimeDefault
    capabilities:
      drop:
      - ALL
  manageTLS: true
  networkPolicy:
    disable: true
  applicationImage: icr.io/automation-saas-platform/access-management/account-iam:20240430132235-main-70f5d498c1867f5da50a2b1bf010eae8e4a4b1c1
  pullPolicy: Always
  replicas: 
  probes:
    startup:
      httpGet:
        scheme: HTTPS
        path: /api/2.0/health/started
        port: 9445
      failureThreshold: 60
      periodSeconds: 10
      timeoutSeconds: 5
    liveness:
      httpGet:
        scheme: HTTPS
        path: /api/2.0/health/liveness
        port: 9445
      timeoutSeconds: 5
      periodSeconds: 10
      failureThreshold: 5
    readiness:
      httpGet:
        scheme: HTTPS
        path: /api/2.0/health/readiness
        port: 9445
      timeoutSeconds: 5
      periodSeconds: 10  
  service:
    type: ClusterIP
    port:  9445
  expose: true
  createKnativeService: false
  resources:
    requests:
      cpu: 300m
      memory: 400Mi
    limits:
      cpu: 1500m
      memory: 800Mi
  volumes:
    - name: account-iam-token
      projected:
        sources:
          - serviceAccountToken:
              audience: openshift
              expirationSeconds: 7200
              path: account-iam-token
        defaultMode: 420
    - name: account-iam-oidc
      secret:
        secretName: account-iam-oidc-client-auth
        defaultMode: 420
    - name: account-iam-cert-bk
      secret:
        secretName: account-iam-svc-tls-cm-autobackup
        defaultMode: 420
    - name: account-iam-okd
      secret:
        secretName: account-iam-okd-auth
        defaultMode: 420
    - name: account-iam-variables
      projected:
        defaultMode: 420
        sources:
          - secret:
              name: account-iam-database-secret
          - secret:
              name: account-iam-mpconfig-secrets
    - name: apiserver-cert
      secret:
        secretName: service-network-serving-signer
        defaultMode: 420
  volumeMounts:
    - name: account-iam-token
      mountPath: /var/run/secrets/tokens
    - name: account-iam-oidc
      readOnly: true
      mountPath: /config/variables/oidc
    - name: account-iam-cert-bk
      readOnly: true
      mountPath: /etc/x509/certs-autobackup
    - name: account-iam-okd
      readOnly: true
      mountPath: /config/variables/okd
    - name: account-iam-variables
      readOnly: true
      mountPath: /config/variables
    - name: apiserver-cert
      mountPath: /var/openshift/apiserver
  env:
    - name: cert_defaultKeyStore
      value: /var/openshift/apiserver/tls.crt
  envFrom:
    - configMapRef:
        name: account-iam-env-configmap-dev
`
