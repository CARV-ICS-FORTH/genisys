kubectl exec -i -t $1 -n $2 -- squeue --j=$3 "-o %.7i %.9P %.8j %.8u %.2t %.10M %.6D %C %N"  | grep -v 'CPUS'