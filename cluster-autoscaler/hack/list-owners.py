#!/usr/bin/env python3

# Copyright 2024 The Kubernetes Authors.
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

""" Python script to list all OWNERS of various parts of Cluster Autoscaler.

Traverse all subdirectories and find OWNERS. This is useful for tagging people
on broad announcements, for instance before a new patch release.
"""

import glob
import yaml
import sys

files = glob.glob('**/OWNERS', recursive=True)
owners = set()

for fname in files:
  with open(fname) as f:
    parsed = yaml.safe_load(f)
    if 'approvers' in parsed and parsed['approvers'] is not None:
      for approver in parsed['approvers']:
        owners.add(approver)
    else:
      print("No approvers found in {}: {}".format(fname, parsed), file=sys.stderr)

for owner in sorted(owners):
  print('@', owner, sep='')
