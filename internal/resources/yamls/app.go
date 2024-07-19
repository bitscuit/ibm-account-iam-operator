package yamls

var APP_SECRETS = []string{
	ClientAuth,
	OKD_Auth,
	DatabaseSecret,
	MpConfig,
}

var APP_STATIC_YAMLS = []string{
	INGRESS,
	EGRESS,
	CONFIG_ENV,
	CONFIG_JWT,
	DB_MIGRATION_MCSPID_SA,
	DB_MIGRATION_MCSPID,
	ACCOUNT_IAM_APP,
}

var ClientAuth = `
kind: Secret
apiVersion: v1
metadata:
  name: account-iam-oidc-client-auth
  namespace: mcsp
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
  annotations:
    argocd.argoproj.io/sync-wave: "0"
data:
  realm: {{ .Realm }}
  client_id: {{ .ClientID }}
  client_secret: {{ .ClientSecret }}
  discovery_endpoint: {{ .DiscoveryEndpoint }}
type: Opaque
`
var OKD_Auth = `
kind: Secret
apiVersion: v1
metadata:
  name: account-iam-okd-auth
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services 
  annotations:
    argocd.argoproj.io/sync-wave: "0"
data:
  user_validation_api_v2: {{ .UserValidationAPIV2 }}
type: Opaque
`

var DatabaseSecret = `
kind: Secret
apiVersion: v1
metadata:
  name: account-iam-database-secret
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
  annotations:
    argocd.argoproj.io/sync-wave: "0"
stringData:
  pg_jdbc_host: common-service-db-rw
  pg_jdbc_port: "5432"
  pg_db_name: account_iam
  pg_db_schema: accountiam
  pg_db_user: user_accountiam
  pg_jdbc_password_jndi: "jdbc/iamdatasource"
  pgPassword: {{ .PGPassword }}
data:
  GLOBAL_ACCOUNT_AUD: {{ .GlobalAccountAud }}
  GLOBAL_ACCOUNT_IDP: {{ .GlobalAccountIDP }}
  GLOBAL_ACCOUNT_REALM: {{ .GlobalRealmValue }}
type: Opaque
`

var MpConfig = `
kind: Secret
apiVersion: v1
metadata:
  name: account-iam-mpconfig-secrets
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
  annotations:
    argocd.argoproj.io/sync-wave: "0"
stringData:
  SRE_MCSP_GROUPS_TOKEN: 
data:
  DEFAULT_AUD_VALUE: {{ .DefaultAUDValue }}
  DEFAULT_IDP_VALUE: {{ .DefaultIDPValue }}
  DEFAULT_REALM_VALUE: {{ .DefaultRealmValue }}
  IBM_VERIFY_URL: 
  CLIENT_ID_FOR_SIUSER: 
  CLIENT_SECRET_FOR_SIUSER: 
  CLIENT_ID_FOR_SERVICEID: 
  CLIENT_SECRET_FOR_SERVICEID: 
type: Opaque
`

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
          port: 5432
`

const CONFIG_ENV = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: account-iam-env-configmap-dev
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

const DB_MIGRATION_MCSPID = `
apiVersion: batch/v1
kind: Job
metadata:
  name: account-iam-db-migration-mcspid
  labels:
    by-squad: mcsp-user-management
    for-product: all
    bcdr-candidate: t
    component-name: iam-services
  annotations:
    argocd.argoproj.io/sync-wave: "2"
    test: mcspid
spec:
  backoffLimit: 3
  activeDeadlineSeconds: 21600
  template:
    metadata:
      labels:
        by-squad: mcsp-user-management
        for-product: all
        name: account-iam
    spec:
      restartPolicy: Never
      containers:
        - name: dbmigrate
          image: icr.io/automation-saas-platform/access-management/account-iam:20240430132235-main-70f5d498c1867f5da50a2b1bf010eae8e4a4b1c1
          envFrom:
            - secretRef:
                name: account-iam-database-secret
          command:
            - /bin/sh
            - '-c'
          args:
            - '/dbmigration/run.sh'
          volumeMounts:
          imagePullPolicy: Always
          resources:
            requests:
              cpu: 100m
              memory: 300Mi
            limits:
              cpu: 500m
              memory: 600Mi
      serviceAccountName: account-iam-migration
      volumes:
        - name: account-iam-token
          projected:
            sources:
              - serviceAccountToken:
                  audience: openshift
                  expirationSeconds: 7200
                  path: account-iam-token
            defaultMode: 420
`

const DB_MIGRATION_MCSPID_SA = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: account-iam-migration
  labels:
    by-squad: mcsp-user-management
    for-product: all
  annotations:
    argocd.argoproj.io/sync-wave: "1"
`

const ACCOUNT_IAM_APP = `
apiVersion: liberty.websphere.ibm.com/v1
kind: WebSphereLibertyApplication
metadata:
  name: account-iam
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
