#!/usr/bin/bash -x

CREDS="./.creds/kafka.client.properties"

cat <<EOF > "$CREDS"
security.protocol=PLAINTEXT
sasl.mechanism=SCRAM-SHA-256
sasl.jaas.config=org.apache.kafka.common.security.scram.ScramLoginModule required username="user1" password="$(kubectl get secret kafka-user-passwords --namespace default -o jsonpath='{.data.client-passwords}' | base64 -d | cut -d , -f 1)";
EOF

kubectl run kafka-client --restart='Never' --image docker.io/bitnami/kafka:3.6.1-debian-11-r4 --namespace default --command -- sleep infinity
kubectl cp --namespace default "$CREDS" kafka-client:/tmp/client.properties
kubectl exec --tty -i kafka-client --namespace default -- bash
