default_registry('ttl.sh/24466dd2-318a-44df-8949-b82b8a64e85a')

docker_build('lumos', '.',
             dockerfile='Dockerfile')

k8s_yaml([
  'dev/k8s/namespace.yaml',
  'dev/k8s/deployment.yaml'
])

allow_k8s_contexts('admin@homelab')

k8s_resource('lumos', port_forwards=[8080])
