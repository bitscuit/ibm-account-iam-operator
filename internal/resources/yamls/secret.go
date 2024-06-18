package yamls

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
  namespace: mcsp
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
  namespace: mcsp
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
  GLOBAL_ACCOUNT_AUD: <>
  GLOBAL_ACCOUNT_IDP: <>
  GLOBAL_ACCOUNT_REALM: <>
type: Opaque
`

var MpConfig = `
kind: Secret
apiVersion: v1
metadata:
  name: account-iam-mpconfig-secrets
  namespace: mcsp
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
