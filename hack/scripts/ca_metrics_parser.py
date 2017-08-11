#!/usr/bin/env python

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the 'License');
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an 'AS IS' BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

'''
This script parses metrics from Cluster Autoscaler e2e tests.
'''
from __future__ import division
from __future__ import print_function
import argparse
import json


class CAMetric(object):

  def __init__(self, function_name):
    self.function_name = function_name
    self.sum = 0.0
    self.average = 0.0
    self.buckets = []
    self.count = 0
    self.upper_bound = 0.0

  def print(self):
    print(self.function_name, '\t', self.sum, '\t', self.count,'\t',  self.avg,
      '\t',  self.upper_bound)
    print(self.buckets)


def print_summary(summary):
  print('function_name\t sum\t count\t avg\t upper_bound')
  print('buckets')
  for metric in summary.values():
    metric.print()


def function_name(sample):
  return sample['metric']['function']


def metric_value(sample):
  return sample['value'][1]


def upper_bound(buckets):
  '''
  Going from the rightmost bucket, find the first one that has some samples
  and return its upper bound.
  '''
  for i in xrange(len(buckets) - 1, -1, -1):
    le, count = buckets[i]
    if i == 0:
      return le
    else:
      le_prev, count_prev = buckets[i-1]
      if count_prev < count:
        return le


def parse_metrics_file(metrics_file):
  '''
  Return interesting metrics for all Cluster Autoscaler functions.

  Merics are stored in a map keyed by function name and are expressed in 
  seconds. They include
  * sum of all samples
  * count of sumples
  * average value of samples
  * upper bound - all collected samples were smaller than this value
  * buckets - list of tuples (# of samples, bucket upper bound)
  '''
  summary = {}
  with open(metrics_file) as metrics_file:
    summary = {}
    metrics = json.load(metrics_file)
    ca_metrics = metrics['ClusterAutoscalerMetrics']

    total_sum = ca_metrics['cluster_autoscaler_function_duration_seconds_sum']
    for sample in total_sum:
      function = function_name(sample)
      summary[function] = CAMetric(function)
      summary[function].sum = float(metric_value(sample))

    count = ca_metrics['cluster_autoscaler_function_duration_seconds_count']
    for sample in count:
      function = function_name(sample)
      summary[function].count = int(metric_value(sample))
      summary[function].avg = summary[function].sum / summary[function].count

    buckets = ca_metrics['cluster_autoscaler_function_duration_seconds_bucket']
    for sample in buckets:
      function = function_name(sample)
      summary[function].buckets.append(
        (float(sample['metric']['le']), int(metric_value(sample))))

    for value in summary.values():
      value.upper_bound = upper_bound(value.buckets)
  return summary


def main():
  parser = argparse.ArgumentParser(description='Parse metrics from Cluster Autoscaler e2e test')
  parser.add_argument('metrics_file', help='File to read metrics from')
  args = parser.parse_args()

  summary = parse_metrics_file(args.metrics_file)  
  print_summary(summary)


if __name__ == '__main__':
  main()
  
