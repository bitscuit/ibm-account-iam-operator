package yamls

var CertRotationYamls = []string{
	CERT_ROTATION_ROLE,
	CERT_ROTATION_ROLE_DEV,
	CERT_ROTATION_RB,
	CERT_ROTATION_RB_DEV,
	CERT_ROTATION_SA,
	CERT_ROTATION_MANAGER,
}

const CERT_ROTATION_MANAGER = `
kind: Deployment
apiVersion: apps/v1
metadata:
  name: iam-cert-rotation-manager
  labels:
    app.kubernetes.io/instance: msp-iam-cert-rotation-ap-dp-001
    bcdr-candidate: t
    by-squad: mcsp-access-management
    component-name: iam-services
    for-product: all
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: iam-cert-rotation-manager
  template:
    metadata:
      creationTimestamp: null
      labels:
        by-squad: mcsp-access-management
        control-plane: iam-cert-rotation-manager
        for-product: all
      annotations:
        kubectl.kubernetes.io/default-container: manager
    spec:
      restartPolicy: Always
      serviceAccountName: msp-iam-cert-rotation-sa
      schedulerName: default-scheduler
      terminationGracePeriodSeconds: 10
      securityContext:
        runAsNonRoot: true
      containers:
        - resources:
            limits:
              cpu: 5m
              memory: 64Mi
            requests:
              cpu: 2m
              memory: 32Mi
          readinessProbe:
            httpGet:
              path: /readyz
              port: 8081
              scheme: HTTP
            initialDelaySeconds: 5
            timeoutSeconds: 1
            periodSeconds: 10
            successThreshold: 1
            failureThreshold: 3
          terminationMessagePath: /dev/termination-log
          name: manager
          command:
            - /manager
          livenessProbe:
            httpGet:
              path: /healthz
              port: 8081
              scheme: HTTP
            initialDelaySeconds: 15
            timeoutSeconds: 1
            periodSeconds: 20
            successThreshold: 1
            failureThreshold: 3
          env:
            - name: WATCH_NAMESPACE
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: metadata.namespace
          securityContext:
            capabilities:
              drop:
                - ALL
            allowPrivilegeEscalation: false
          imagePullPolicy: Always
          terminationMessagePolicy: File
          image: >-
            icr.io/automation-saas-platform/access-management/iam-cert-rotation:20240306103454-main-86f22aa63ce252c4add52c8c7bf11ff24c430764
      serviceAccount: msp-iam-cert-rotation-sa
      dnsPolicy: ClusterFirst
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 25%
      maxSurge: 25%
  revisionHistoryLimit: 10
  progressDeadlineSeconds: 600
`

const CERT_ROTATION_SA = `
kind: ServiceAccount
apiVersion: v1
metadata:
  name: msp-iam-cert-rotation-sa
  labels:
    app.kubernetes.io/instance: msp-iam-cert-rotation-ap-dp-001
    bcdr-candidate: t
    by-squad: mcsp-access-management
    component-name: iam-services
    for-product: all
`

const CERT_ROTATION_ROLE = `
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: msp-iam-cert-rotation-role
  labels:
    app.kubernetes.io/instance: msp-iam-cert-rotation-ap-dp-001
    bcdr-candidate: t
    by-squad: mcsp-access-management
    component-name: iam-services
    for-product: all
rules:
  - verbs:
      - get
      - list
      - watch
      - create
      - update
      - patch
      - delete
    apiGroups:
      - ''
    resources:
      - secrets
  - verbs:
      - get
      - list
      - watch
      - create
      - update
      - patch
      - delete
    apiGroups:
      - coordination.k8s.io
    resources:
      - leases
`

const CERT_ROTATION_RB = `
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: msp-iam-cert-rotation-rb
  labels:
    app.kubernetes.io/instance: msp-iam-cert-rotation-ap-dp-001
    bcdr-candidate: t
    by-squad: mcsp-access-management
    component-name: iam-services
    for-product: all
subjects:
  - kind: ServiceAccount
    name: msp-iam-cert-rotation-sa
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: msp-iam-cert-rotation-role
`

const CERT_ROTATION_RB_DEV = `
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: msp-iam-cert-rotation-rb-development
  labels:
    app.kubernetes.io/instance: msp-iam-cert-rotation-ap-dp-001
    by-squad: mcsp-access-management
    for-product: all
subjects:
  - kind: ServiceAccount
    name: msp-iam-cert-rotation-sa
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: msp-iam-cert-rotation-role-development
`

const CERT_ROTATION_ROLE_DEV = `
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: msp-iam-cert-rotation-role-development
  labels:
    app.kubernetes.io/instance: msp-iam-cert-rotation-ap-dp-001
    by-squad: mcsp-access-management
    for-product: all
rules:
  - verbs:
      - use
    apiGroups:
      - security.openshift.io
    resources:
      - securitycontextconstraints
`
