#!/usr/bin/env python

# Copyright 2017 The Kubernetes Authors.
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

'''
This script breaks a given GCE MIG to simulate zone failure or similar disaster
scenario for testing purposes.

It works by polling `gcloud compute instances list` and adding iptables rules
on master to block ip addresses of instances, whose name matches pattern.
The script runs in endless until you kill it with signal (ctrl-c?) and than
it cleans up (remove iptables rules it added) before exiting.

Run with -e flag to break existing nodes in the node group and -u to break
new nodes added after the script was started. You're free to use both this
flags together to break all nodes.

Messing with iptables rules on master is obviously unsafe and can potentially
lead to completely breaking your cluster!
'''
from __future__ import print_function


import argparse
import atexit
import collections
import re
import subprocess
import sys
import time


InstanceInfo = collections.namedtuple("InstanceInfo", 'name ip')


def get_instances(master, ng):
    '''Poll instances list and parse result to list of InstanceInfo structs'''
    raw = subprocess.check_output(['gcloud', 'compute', 'instances', 'list'])
    first = True
    result = []
    for l in raw.splitlines():
        if first:
            first = False
            continue
        parts = l.split()
        name = parts[0]
        if not name.startswith(ng):
            continue
        ips = []
        for p in parts[1:]:
          if re.match('([0-9]{1,3}\.){3}[0-9]{1,3}', p):
              ips.append(p)
        # XXX: A VM has showed up, but it doesn't have internal and external ip
        # yet, let's just pretend we haven't seen it yet
        if len(ips) < 2:
              continue
        info = InstanceInfo(name, ips)
        result.append(info)
    return result


def break_node(master, instance, broken_ips, verbose):
    '''Add iptable rules to drop packets coming from ips used by a give node'''
    print('Breaking node {}'.format(instance.name))
    for ip in instance.ip:
        if verbose:
            print('Blocking ip {} on master'.format(ip))
        subprocess.call(['gcloud', 'compute', 'ssh', master, '--', 'sudo iptables -I INPUT 1 -p tcp -s {} -j DROP'.format(ip)])
        broken_ips.add(ip)


def run(master, ng, existing, upcoming, max_nodes_to_break, broken_ips, verbose):
    '''
    Poll for new nodes and break them as required.

    Runs an endless loop.
    '''

    # can't assign to local variable from nested function in python 2
    # but can mutate a list (standard hack)
    broken = [0]

    def maybe_break_node(*args, **kwargs):
        if max_nodes_to_break >= 0 and broken[0] >= max_nodes_to_break:
            if verbose:
                print('Maximum number of instances already broken, will not break {}'.format(args[1]))
        else:
            break_node(*args, **kwargs)
            broken[0] += 1

    instances = get_instances(master, ng)
    known = set()
    for inst in instances:
        if existing:
            maybe_break_node(master, inst, broken_ips, verbose)
        known.add(inst.name)
    while True:
        instances = get_instances(master, ng)
        for inst in instances:
            if inst.name in known:
                continue
            if verbose:
                print('New instance observed: {}'.format(inst.name))
            if upcoming:
                maybe_break_node(master, inst, broken_ips, verbose)
            known.add(inst.name)
        time.sleep(5)


def clean_up(master, broken, verbose):
    '''
    Clean up iptable rules created by this script.

    WARNING: this just deletes top N rules if you've added some rules to the
    top of INPUT chain while this was running you will suffer.
    '''
    if verbose:
        print('Cleaning up top {} iptable rules'.format(len(broken)))
    for i in xrange(len(broken)):
        subprocess.call(['gcloud', 'compute', 'ssh', master, '--', 'sudo iptables -D INPUT 1'])


def main():
    parser = argparse.ArgumentParser(description='Break all existing and/or upcoming node in a MIG')
    parser.add_argument('master_name', help='name of kubernetes master (will be used with gcloud)')
    parser.add_argument('node_group_name', help='name of node group to break')
    parser.add_argument('-e', '--existing', help='break existing nodes (they will become unavailable)', action='store_true')
    parser.add_argument('-u', '--upcoming', help='break any new nodes added to this node group (they will not register at all)', action='store_true')
    parser.add_argument('-m', '--max-nodes-to-break', help='break at most a given number of nodes', type=int, default=-1)
    parser.add_argument('-v', '--verbose', action='store_true')
    parser.add_argument('-y', '--yes', action='store_true')
    args = parser.parse_args()

    if not args.existing and not args.upcoming:
        print('At least one of --existing or --upcoming must be specified')
        return

    if not args.yes:
        print('Running this script will break nodes in your cluster for testing purposes.')
        print('The nodes may or may not recover after this. Your whole cluster may be broken.')
        print('DO NOT RUN THIS SCRIPT ON PRODUCTION CLUSTER.')
        print('Do you want to proceed? (anything but y stops the script)')
        user_ok = sys.stdin.read(1)
        if user_ok.upper() != 'Y':
            return

    broken = set()
    atexit.register(clean_up, args.master_name, broken, args.verbose)
    run(args.master_name, args.node_group_name, args.existing, args.upcoming, args.max_nodes_to_break, broken, args.verbose)


if __name__ == '__main__':
    main()
