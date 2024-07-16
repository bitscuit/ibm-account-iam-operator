package yamls

const UTILS_JOB_ROLE = `
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: mcsp-utils
rules:
  - verbs:
      - get
      - list
      - watch
    apiGroups:
      - ''
    resources:
      - secrets
	  - configmaps
	  - pods
  - verbs:
      - create
      - update
      - delete
    apiGroups:
      - ''
    resources:
      - secrets
`

const UTILS_JOB_SA = `
kind: ServiceAccount
apiVersion: v1
metadata:
  name: mcsp-utils
`

const UTILS_JOB_RB = `
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: mcsp-utils
subjects:
  - kind: ServiceAccount
    name: mcsp-utils
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: mcsp-utils
`
