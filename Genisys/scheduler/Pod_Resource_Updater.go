// Copyright [2021] [FORTH-ICS]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"strconv"
	"time"
)

// Resource_Bandwidth_Updater_Nodes() iterates over each Cluster's Node, for each Node it computes the Sum of all the consumed resources of the Deployments placed on the spesific Node
// and assigns these values to an Resource Allocation Map
// Resource_Bandwidth_Updater_Nodes() is necessary when updating the Nodes allocations
func Resource_Bandwidth_Updater_Nodes() {

	for {
		select {
		case <-done:
			return
		case _ = <-ticker.C:
			Wait_For_Nodes_Resources_Update()
			Timer_Node_Updater = 1
			Node_List, _ := Get_All_Nodes()

			for _, Node := range Node_List.Items {
				Node_Net_Limit[Node.ObjectMeta.Name] = 0
				Node_IO_Limit[Node.ObjectMeta.Name] = 0
				Node_Cpu_Limit[Node.ObjectMeta.Name] = 0
			}

			for _, namespace := range Namespaces {
				Pods, err := Get_Pods_Of_Namespace(namespace)
				if err != nil {
					continue
				}
				for it := 0; it < len(Pods.Items); it++ {

					if Pods.Items[it].Spec.NodeName == "" || Pods.Items[it].Status.Phase == "Terminating" {
						continue
					}

					Deployment, err := Get_Deployment_Of_Pod(&Pods.Items[it])
					if err != nil {
						continue
					}

					Deployment_Name, err := Get_Deployment_Name_Of_Pod(&Pods.Items[it])
					if err != nil {
						continue
					}

					//Wait_For_Deployment_Resources_Update()
					Some_Map_Mutex.Lock()

					//Update Node saturated net bandwidth
					//Deployment_Net_Limit[Deployment_Name+namespace], _ = strconv.ParseInt(Deployment.ObjectMeta.Annotations["net-bandwidth"], 10, 64)

					//Node_Net_Limit[Pods.Items[it].Spec.NodeName] = Node_Net_Limit[Pods.Items[it].Spec.NodeName] + Deployment_Net_Limit[Deployment_Name+namespace]

					//Update Node saturated io
					//Deployment_IO_Limit[Deployment_Name+namespace], _ = strconv.ParseInt(Deployment.ObjectMeta.Annotations["io-bandwidth"], 10, 64)

					//Node_IO_Limit[Pods.Items[it].Spec.NodeName] = Node_IO_Limit[Pods.Items[it].Spec.NodeName] + Deployment_IO_Limit[Deployment_Name+namespace]

					//Update Node saturated cpu
					Deployment_Cpu_Limit[Deployment_Name+namespace], _ = strconv.ParseInt(Deployment.ObjectMeta.Annotations["cpu-bandwidth"], 10, 64)

					Node_Cpu_Limit[Pods.Items[it].Spec.NodeName] = Node_Cpu_Limit[Pods.Items[it].Spec.NodeName] + Deployment_Cpu_Limit[Deployment_Name+namespace]

					Some_Map_Mutex.Unlock()

				}
			}

			Timer_Node_Updater = 0
			time.Sleep(100 * time.Millisecond) //changes 17/5/2020
		}
	}

}

// In order to ensure that that no POD is scheduled during the Resource_Bandwidth_Updater_Nodes proccess Wait_For_Nodes_Resources_Update_To_Start() is used.
// Wait_For_Nodes_Resources_Update() stalls while te Nodes resources are updated.

func Wait_For_Nodes_Resources_Update_To_Start() {
	for Timer_Node_Updater == 0 {
		time.Sleep(1 * time.Millisecond) //experimental
	}
}

// In order to ensure that that no POD is scheduled during the Resource_Bandwidth_Updater_Nodes proccess Wait_For_Nodes_Resources_Update() is used.
// Wait_For_Nodes_Resources_Update() stalls while te Nodes resources are updated.

func Wait_For_Nodes_Resources_Update() {
	for Timer_Node_Updater == 1 {
		time.Sleep(1 * time.Millisecond) //experimental
	}
}

// Wait_For_Deployment_Resources_Update_To_Start() Waits for Deployment Resources Update
func Wait_For_Deployment_Resources_Update_To_Start() {
	for Timer_Deployment_Updater == 0 {
		time.Sleep(1 * time.Millisecond) //avoid spinning
	}
}

// Wait_For_Deployment_Resources_Update() follows the same pattern except for Deployment Update proccedure
func Wait_For_Deployment_Resources_Update() {
	for Timer_Deployment_Updater == 1 {
		time.Sleep(1 * time.Millisecond) //avoid spinning
	}
}

func Resources_Bandwidth_Printer_Nodes() {

	for {
		select {
		case <-done:
			return
		case _ = <-ticker.C:
			Node_List, _ := Get_All_Nodes()
			Wait_For_Nodes_Resources_Update()
			fmt.Println("\n													######################################################")
			fmt.Println("													ALLOCATED RESOURCES FOR OF EACH NODE")
			fmt.Println("													######################################################")
			for _, Node := range Node_List.Items {
				if Check_If_Node_Is_Included(&Node) == false || Check_If_Node_Is_Excluded(&Node) == true {
					continue
				}
				s := fmt.Sprintf("%d", Node_Net_Limit[Node.ObjectMeta.Name]/8000)
				fmt.Println("													" + Node.ObjectMeta.Name + " bandwidth limit (MB/s) : " + s)
				s = fmt.Sprintf("%d", Node_IO_Limit[Node.ObjectMeta.Name]/1000000)
				fmt.Println("													" + Node.ObjectMeta.Name + " io limit (MB/s) : " + s)
				s = fmt.Sprintf("%d", Node_Cpu_Limit[Node.ObjectMeta.Name]/1000)
				fmt.Println("													" + Node.ObjectMeta.Name + " cpu limit : " + s)
				fmt.Println("\n")
			}
			fmt.Println("\n\n")
			time.Sleep(10000 * time.Millisecond)
		}
	}
}

// Resources_Bandwidth_Updater_Deployments() iterates over each Cluster's Deployment, from each Deployment
// it reads the specified resource limits from the deployments METADATA and updates the Deployment resource limit map
// as a next step it procceeds to enforce the deployment's resource limits using  the CGROUPS driver
func Resources_Bandwidth_Updater_Deployments() {

	for {
		select {
		case <-done:
			return
		case _ = <-ticker.C:
			for _, namespace := range Namespaces {
				Pods, _ := Get_Pods_Of_Namespace(namespace)
				Timer_Deployment_Updater = 1

				for it := 0; it < len(Pods.Items); it++ {

					Deployment, err := Get_Deployment_Of_Pod(&Pods.Items[it])
					if err != nil {
						continue
					}

					Deployment_Name, err := Get_Deployment_Name_Of_Pod(&Pods.Items[it])
					if err != nil {
						continue
					}

					if err != nil {
						continue
					}

					if Pods.Items[it].Spec.SchedulerName != *Scheduler_Name {
						continue
					}

					Some_Map_Mutex.Lock()

					Deployment_Net_Limit[Deployment_Name+namespace], _ = strconv.ParseInt(Deployment.ObjectMeta.Annotations["net-bandwidth"], 10, 64)
					Deployment_IO_Limit[Deployment_Name+namespace], _ = strconv.ParseInt(Deployment.ObjectMeta.Annotations["io-bandwidth"], 10, 64)
					Deployment_Cpu_Limit[Deployment_Name+namespace], _ = strconv.ParseInt(Deployment.ObjectMeta.Annotations["cpu-bandwidth"], 10, 64)

					Some_Map_Mutex.Unlock()

					//Update nwtwork bandwidth for each Pod

					//Network_Limit := fmt.Sprintf("%d", Deployment_Net_Limit[Deployment_Name+namespace])
					//Shellout("kubectl exec -it " + Pods.Items[it].ObjectMeta.Name + " -n " + namespace + " -- bash -c \"./net-limiter " + Network_Limit + " " + Network_Limit + "\"")

					//Update io bandwidth for each Pod

					//IO_Limit := fmt.Sprintf("%d", Deployment_IO_Limit[Deployment_Name+namespace])

					//Shellout("kubectl exec -it " + Pods.Items[it].ObjectMeta.Name + " -n " + namespace + " -- sudo bash -c 'echo \"8:0 " + IO_Limit + "\" > /sys/fs/cgroup/blkio/blkio.throttle.write_bps_device'")

					//Update cpu for each Pod

					Cpu_Limit := fmt.Sprintf("%d", Deployment_Cpu_Limit[Deployment_Name+namespace]*100) //multiply with 100 to be cgroup ready

					Shellout("./cpu_limiter_datacenter.sh " + Pods.Items[it].ObjectMeta.Name + " " + namespace + " " + Cpu_Limit + " " + strconv.FormatInt((Node_Capacity_Of_Node(Pods.Items[it].Spec.NodeName)/1000)-1, 10))
				}
				Timer_Deployment_Updater = 0
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// Resources_Cpu_Limiter_Virtual_Clusters() iterates over each Dummy Pod computes the actual resources used by the corresponding Slurm Job for each Virtual Cluste's Node,
// and enforces the resource limit using the CGROUPS driver
func Resources_Cpu_Limiter_Virtual_Clusters() {

	var Cpu_Pod_Limit int64
	var Cpu_Limit_Virtual_Node int64
	Cpu_Pod_Limit = 0

	for {
		select {
		case <-done:
			return
		case _ = <-ticker.C:
			for _, namespace := range Namespaces {
				Slurm_Pods, _ := Get_Virtual_cluster_Pods(namespace)
				for it1 := 0; it1 < len(Slurm_Pods.Items); it1++ {
					//reset Cpu limit for each Virtual Slurm Node
					Cpu_Limit_Virtual_Node = 0
					dummy_pods, _ := Get_Dummy_MPI_Pods(namespace)
					for it2 := 0; it2 < len(dummy_pods.Items); it2++ {
						if Slurm_Pods.Items[it1].Spec.NodeName == dummy_pods.Items[it2].Spec.NodeName {
							Deployment, err := Get_Deployment_Of_Pod(&dummy_pods.Items[it2])
							if err != nil {
								continue
							}
							if Deployment.ObjectMeta.Annotations["type"] == "MPI-JOB" {
								Cpu_Pod_Limit, _ = strconv.ParseInt(Deployment.ObjectMeta.Annotations["cpu-bandwidth"], 10, 64)
								Cpu_Pod_Limit = Cpu_Pod_Limit * 100
								Cpu_Limit_Virtual_Node += Cpu_Pod_Limit
							}
						}
						Deployment, err := Get_Deployment_Of_Pod(&dummy_pods.Items[it2])
						if err != nil {
							continue
						}
						if dummy_pods.Items[it2].Status.Phase == "Running" {
							Shellout("./slurm_priority " + Slurm_Pods.Items[it1].ObjectMeta.Name + " " + namespace + " " + Deployment.ObjectMeta.Name)
						}
						if it2 == len(dummy_pods.Items)-1 {
							Cpu_Limit := fmt.Sprintf("%d", 6200000) //Cpu_Limit_Virtual_Node) //hotfix
							Shellout("./cpu_limiter_slurm.sh " + Slurm_Pods.Items[it1].ObjectMeta.Name + " " + namespace + " " + Cpu_Limit + " " + strconv.FormatInt((Node_Capacity_Of_Node(Slurm_Pods.Items[it1].Spec.NodeName)/1000)-1, 10))
						}
					}
				}
			}
		}
	}
}
