#!/bin/sh
d=${1?Pod_Name}
e=${2:-namespace}
u=${3:-cpu_limit}

#echo 0 > /sys/fs/cgroup/cpuset/cpuset.cpus
kubectl exec -it $d -n $e -- bash -c 'echo '$u' > /sys/fs/cgroup/cpu/cpu.cfs_quota_us'         
kubectl exec -it $d -n $e -- sudo sh -c 'echo '$u' > /sys/fs/cgroup/cpu/cpu.cfs_quota_us'
