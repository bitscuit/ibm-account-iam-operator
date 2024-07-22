package yamls

var IMConfigYamls = []string{
	IM_CONFIG_ROLE,
	IM_CONFIG_ROLE_BINDING,
	IM_CONFIG_SA,
	IM_CONFIG_JOB,
}

var IM_CONFIG_JOB = `
apiVersion: batch/v1
kind: Job
metadata:
  name: mcsp-im-config-job
  labels:
    app: mcsp-im-config-job
spec:
  template:
    metadata:
      labels:
        app: mcsp-im-config-job
    spec:
      containers:
      - name: mcsp-im-config-job
        image: docker-na-public.artifactory.swg-devops.com/hyc-cloud-private-scratch-docker-local/ibmcom/mcsp-im-config-job-amd64:5796e4d
        command: ["./mcsp-im-config-job"]
        imagePullPolicy: Always
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
              - ALL
          runAsNonRoot: true
          seccompProfile:
            type: RuntimeDefault
        env:
          - name: LOG_LEVEL
            value: debug
          - name: NAMESPACE
            value: {{ .AccountIAMNamespace }}
          - name: IM_HOST_URL
            value: {{ .DefaultIDPValue }}
          - name: ACCOUNT_IAM_URL
            value: {{ .AccountIAMURL }}
      serviceAccountName: mcsp-im-config-sa
      restartPolicy: OnFailure
`

var IM_CONFIG_SA = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mcsp-im-config-sa
  labels:
    app: mcsp-im-config-sa
`

var IM_CONFIG_ROLE = `
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: mcsp-im-config-role
  labels:
    app: mcsp-im-config-role
rules:
  - apiGroups: [""]
    resources: ["pods", "configmaps", "secrets"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    resourceNames: ["mcsp-im-integration-api-key"]
    verbs: ["create", "update", "delete"]
`

var IM_CONFIG_ROLE_BINDING = `
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: mcsp-im-config-rb
  labels:
    app: mcsp-im-config-rb
subjects:
  - kind: ServiceAccount
    name: mcsp-im-config-sa
roleRef:
  kind: Role
  name: mcsp-im-config-role
  apiGroup: rbac.authorization.k8s.io
`
