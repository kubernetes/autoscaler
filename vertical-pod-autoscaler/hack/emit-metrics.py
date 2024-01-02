#!/usr/bin/env python3

# Copyright 2023 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script runs as a pod, scanning the cluster for other pods.
# It then publishes fake metrics (Gaussian by --mean_* and --stddev_*)
# for each pod into a Prometheus Pushgateway.  The Prometheus instance
# connected to that Pushgateway can have a Prometheus Adapter connected
# to it that serves as an External Metrics Provider to Kubernetes.

import argparse
import base64
from collections import defaultdict
from kubernetes import client, config
import math
import random
import re
import requests
import sys
import time
import urllib.parse
import pprint

def parse_arguments():
    parser = argparse.ArgumentParser(description='')
    parser.add_argument('--dest', type=str, default='pushservice')
    parser.add_argument('--mean_cpu', type=int, default='1000', help='Mean millicores for cpu.')
    parser.add_argument('--mean_mem', type=int, default='128', help='Mean megabytes for memory.')
    parser.add_argument('--stddev_cpu', type=int, default=150, help='Standard deviation for cpu.')
    parser.add_argument('--stddev_mem', type=int, default=15, help='Standard deviation for mem.')
    parser.add_argument('--sleep_sec', type=int, default=30, help='Delay between metric-sends, in seconds.')
    parser.add_argument('-t','--tags', action='append', nargs=2, metavar=('key','value'), default=[['data', 'emit-metrics']],
                        help='Additional tags to attach to metrics.')
    parser.add_argument('--namespace_pattern', default='monitoring', help='Regex to match namespace names.')
    parser.add_argument('--pod_pattern', default='prometheus-[0-9a-f]{9}-[0-9a-z]{5}', help='Regex to match pod names.')
    parser.add_argument('--all', default=False, action='store_true', help='Write metrics for all pods.')
    parser.add_argument('--job', default='emit-metrics', help='Job name to submit under.')
    return parser.parse_args()

def safestr(s):
    '''Is s a URL-safe string?'''
    return s.strip('_').isalnum()

def urlify(key, value):
    replacements = { ".": "%2E", "-": "%2D" }
    def encode(s):
        s = urllib.parse.quote(s, safe='')
        for c,repl in replacements.items():
            s = s.replace(c, repl)
        return s
    if safestr(key) and safestr(value):
        return f"{key}/{value}"
    elif len(value) == 0:
        # Possibly encode the key using URI encoding, but
        # definitely use base64 for the value.
        return encode(key)+"@base64/="
    else:
        return f"{encode(key)}/{encode(value)}"

def valid_key(key):
    invalid_char_re = re.compile(r'.*[./-].*')
    invalid_keys = set(["pod-template-hash", "k8s-app", "controller-uid",
                        "controller-revision-hash", "pod-template-generation"])
    return (key not in invalid_keys) and invalid_char_re.match(key) == None

def send_metrics(args, job, path, cpuval, memval):
    cpuval = cpuval / 1000.0  # Scale from millicores to cores
    payload = f"cpu {cpuval:.3f}\nmem {memval:d}.0\n"
    path_str = '/'.join([urlify(key,value) for key, value in path.items()])
    url = f'http://{args.dest}/metrics/job/{job}/namespace/{path["namespace"]}/{path_str}'
    response = requests.put(url=url, data=bytes(payload, 'utf-8'))
    if response.status_code != 200:
        print (f"Writing to {url}.\n>> Got {response.status_code}: {response.reason}, {response.text}\n>> Dict was:")
        pprint.pprint(path)
    else:
        print (f"Wrote to {url}: {payload}")
    sys.stdout.flush()

def main(args):
    print (f"Starting up.")
    sys.stdout.flush()
    pod_name_pattern = re.compile(args.pod_pattern)
    namespace_name_pattern = re.compile(args.namespace_pattern)
    try:
        config.load_kube_config()
    except:
        config.load_incluster_config()
    v1 = client.CoreV1Api()
    print (f"Initialized.  Sleep interval is for {args.sleep_sec} seconds.")
    sys.stdout.flush()
    pod_cache = dict()
    while True:
        skipped_keys= set()
        time.sleep(args.sleep_sec)
        pods = v1.list_pod_for_all_namespaces(watch=False)
        all = 0
        found = 0
        for pod in pods.items:
            all += 1
            job = args.job
            if args.all or (namespace_name_pattern.match(pod.metadata.namespace) and pod_name_pattern.match(pod.metadata.name)):
                # Get container names and send metrics for each.
                key = f"{pod.metadata.namespace}/{pod.metadata.name}"
                if key not in pod_cache:
                    v1pod = v1.read_namespaced_pod(pod.metadata.name, pod.metadata.namespace, pretty=False)
                    pod_cache[key] = v1pod
                else:
                    v1pod = pod_cache[key]
                containers = [ c.name for c in v1pod.spec.containers ]
                found += 1
                path = { "kubernetes_namespace": pod.metadata.namespace,
                         "kubernetes_pod_name": pod.metadata.name,
                         "pod": pod.metadata.name,
                         "namespace": pod.metadata.namespace}
                # Append metadata to the data point, add the labels second to overwrite annotations on
                # conflict
                try:
                    if v1pod.metadata.annotations:
                        for k,v in v1pod.metadata.annotations.items():
                            if valid_key(k):
                                path[k] = v
                            else:
                                skipped_keys |= set([k])
                    if v1pod.metadata.labels:
                        for k,v in v1pod.metadata.labels.items():
                            if valid_key(k):
                                path[k] = v
                            else:
                                skipped_keys |= set([k])
                except ValueError as err:
                    print (f"{err} on {v1pod.metadata} when getting annotations/labels")
                if "job" in path:
                    job = path["job"]
                for container in containers:
                    cpuval = random.normalvariate(args.mean_cpu, args.stddev_cpu)
                    memval = random.normalvariate(args.mean_mem, args.stddev_mem)
                    path['name'] = container
                    path['container'] = container
                    send_metrics(args, job, path, math.floor(cpuval), math.floor(memval * 1048576.0))
        print(f"Found {found} out of {all} pods.  Skipped keys:")
        pprint.pprint(skipped_keys)

if __name__ == '__main__':
    main(parse_arguments())
