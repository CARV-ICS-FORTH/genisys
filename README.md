# Genisys Kubernetes scheduler

Genisys is a custom Kubernetes scheduler that distinguishes between “HPC” and “Data Center” services (typical Kubernetes deployments that run in other containers), in order to apply different allocation policies and maximize overall usage. In cases where HPC workloads do not consume all node-local resources, Genisys colocates data center services, while constantly satisfying their user-defined performance targets (as Skynet does). Therefore, HPC and data center workloads execute transparently on the same infrastructure, achieving high levels of CPU utilization.

To run HPC applications in Kubernetes, we introduce the concept of the "Virtual Cluster", as a group of multiple container instances that function as a unified cluster environment from the user’s perspective. Each node in a "Virtual Cluster" embeds all necessary libraries and utilities, as well as a private Slurm deployment offering Infiniband support.

To schedule and place workloads across multiple "Virtual Cluster" and prevent the interference introduced by overlapping jobs, we have modified the Slurm controller’s placement mechanism to delegate all respective decisions to Genisys. Genisys in this scheme is the central authority that has the full knowledge of the cluster’s current resource allocations and acts as a global coordinator for new requests.

For each Slurm job initiated by the user, the modified Slurm creates a resource request for Genisys. When Genisys allocates the required resources from Kubernetes it returns a list with the suitable nodes for the job to run back to Slurm. As a next step Slurm starts the job in the corresponding containers.

For more information, consult the following publications:
```
"Virtual Clusters: Isolated, Containerized HPC Environments in Kubernetes”,
George Zervas, Antony Chazapis, Yannis Sfakianakis, Christos Kozanitis, and Angelos Bilas,
Proceedings of the 17th Workshop on Virtualization in High-Performance Cloud Computing (VHPC’22),
Hamburg, Germany, June 2022.

"Skynet: Performance-driven Resource Management for Dynamic Workloads",
Yannis Sfakianakis, Manolis Marazakis, and Angelos Bilas,
Proceedings of the 2021 IEEE 14th International Conference on Cloud Computing (CLOUD 2021),
Virtual Conference, September 2021.
```

Installation instructions are in [INSTALL](INSTALL.md).

## Acknowledgements

This project has received funding from the European Union’s Horizon 2020 research and innovation programme under grant agreement No 825061 (EVOLVE - [website](https://www.evolve-h2020.eu), [CORDIS](https://cordis.europa.eu/project/id/825061)).
