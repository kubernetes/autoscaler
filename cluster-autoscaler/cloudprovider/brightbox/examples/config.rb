def config
  {
    cluster_name: 'kubernetes.cluster.local',
    image: 'brightbox/cluster-autoscaler-brightbox',
    secret: 'brightbox-credentials'
  }
end

def output(config)
  { 'autoDiscovery' => { 'clusterName' => config[:cluster_name] },
    'cloudProvider' => 'brightbox',
    'image' =>
    { 'repository' => config[:image],
      'tag' => ENV['TAG'],
      'pullPolicy' => 'Always' },
    'tolerations' =>
    [
      { 'effect' => 'NoSchedule', 'key' => 'node-role.kubernetes.io/master' },
      { 'operator' => 'Exists', 'key' => 'CriticalAddonsOnly' }
    ],
    'extraArgs' =>
    { 'v' => (ENV['TAG'] == 'dev' ? 4 : 2).to_s,
      'stderrthreshold' => 'info',
      'logtostderr' => true,
      'cluster-name' => config[:cluster_name],
      'skip-nodes-with-local-storage' => true },
    'podAnnotations' =>
    { 'prometheus.io/scrape' => 'true', 'prometheus.io/port' => '8085' },
    'rbac' => { 'create' => true },
    'resources' =>
    { 'limits' => { 'cpu' => '100m', 'memory' => '300Mi' },
      'requests' => { 'cpu' => '100m', 'memory' => '300Mi' } },
    'envFromSecret' => config[:secret],
    'priorityClassName' => 'system-cluster-critical',
    'dnsPolicy' => 'Default' }
end

require 'yaml'
STDOUT << output(config).to_yaml
