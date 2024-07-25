package yamls

const OperandRequest = `
apiVersion: operator.ibm.com/v1alpha1
kind: OperandRequest
metadata:
  name: ibm-user-management-operator-deps
spec:
  requests:
    - operands:
        - name: ibm-im-operator
      registry: common-service
`
