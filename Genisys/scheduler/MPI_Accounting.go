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

	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
)

// Mpi_Job_Exists_Check() checks if MPI dummy Containers are located on a specified Node

func Mpi_Job_Exists_Check(Node v1.Node) int {

	for _, namespace := range Namespaces {
		Pods, _ := Get_Pods_Of_Namespace(namespace)

		for it := 0; it < len(Pods.Items); it++ {
			if (Pods.Items[it].Status.Phase != "Running" && Pods.Items[it].Status.Phase != "ContainerCreating") || Pods.Items[it].Status.Phase == "Terminating" {
				continue
			}
			if Pods.Items[it].Spec.NodeName == Node.ObjectMeta.Name {
				Deployment, err := Get_Deployment_Of_Pod(&Pods.Items[it])
				if err != nil {
					continue
				}
				if Deployment.ObjectMeta.Annotations["type"] == "MPI-JOB" {
					return 1
				}
			}
		}
	}
	return 0
}

// Mpi_Job_Exists_Check() checks if MPI dummy Containers of the same Deployment are located on a specified Node

func Mpi_Job_Self_Exists_Check(Deployment_Name_Input string, Node v1.Node, Namespace string) int {

	Pods, _ := Get_Pods_Of_Namespace(Namespace)

	for it := 0; it < len(Pods.Items); it++ {
		if (Pods.Items[it].Status.Phase != "Running" && Pods.Items[it].Status.Phase != "ContainerCreating") || Pods.Items[it].Status.Phase == "Terminating" {
			continue
		}
		if Pods.Items[it].Spec.NodeName == Node.ObjectMeta.Name {
			Deployment, err := Get_Deployment_Of_Pod(&Pods.Items[it])
			if err != nil {
				continue
			}
			Deployment_Name, err := Get_Deployment_Name_Of_Pod(&Pods.Items[it])
			if err != nil {
				continue
			}
			if Deployment.ObjectMeta.Annotations["type"] == "MPI-JOB" && Deployment_Name_Input == Deployment_Name {
				return 1
			}
		}
	}

	return 0
}

// Data_Center_Job_Exists_Check() checks if a Data-Center type container exists on a specified Node

func Data_Center_Job_Exists_Check(Node v1.Node) int {

	for _, namespace := range Namespaces {

		Pods, _ := Get_Pods_Of_Namespace(namespace)

		for it := 0; it < len(Pods.Items); it++ {
			if (Pods.Items[it].Status.Phase != "Running" && Pods.Items[it].Status.Phase != "ContainerCreating") || Pods.Items[it].Status.Phase == "Terminating" {
				continue
			}
			if Pods.Items[it].Spec.NodeName == Node.ObjectMeta.Name {
				Deployment, err := Get_Deployment_Of_Pod(&Pods.Items[it])
				if err != nil {
					continue
				}
				if Deployment.ObjectMeta.Annotations["type"] == "DATACENTER-JOB" {
					return 1
				}
			}
		}
	}
	return 0
}

//Check_If_MPI_Job_Is_Scheduled checks if any MPI Job is on Placement phase
func Check_If_MPI_Job_Is_Scheduled() int32 {
	Pods, _ := Get_Dummy_MPI_Pods_All_Namespaces()
	for it := 0; it < len(Pods.Items); it++ {
		Deployment, _ := Get_Deployment_Of_Pod(&Pods.Items[it])
		if Deployment.Status.ReadyReplicas != Deployment.Status.Replicas && Deployment.ObjectMeta.Annotations["scheduled"] == "on" {
			fmt.Println("Another Deployment is being scheduled Right now !")
			return 1
		}
	}
	return 0
}

// Mpi_Job_Check_If_It_Fits() iterates over the Clusters Nodes in order to check if an MPI deployment can fit onto the Cluster.
// If enough resources are available it returns 1, in any other cases returns 0
// Used before the deployment of each Slurm Job in order to check if it fits

func Mpi_Job_Check_If_It_Fits(Pod *v1.Pod) int32 { //bug when creating mpi Deployment at the same time with the deletion of another Deployment

	var replica_fits int
	var node_cpu_limit_temporary = make(map[string]int64)
	var node_has_mpi = make(map[string]string)
	var node_has_data_center = make(map[string]string)
	var Replicas int32

	if Check_If_MPI_Job_Is_Scheduled() == 1 || Check_If_Data_Center_Service_Is_Scheduled() == 1 {
		return 0
	}

	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}

	Deployments := Client_Set.AppsV1().Deployments(Pod.ObjectMeta.Namespace)
	Deployment, err := Get_Deployment_Of_Pod(Pod)
	if err != nil {
		return 0
	}
	if Deployment.ObjectMeta.Annotations["type"] != "MPI-JOB" {
		return 0
	}
	Replicas = Deployment.Status.Replicas

	fmt.Println("Checking for ", Replicas, " Replicas")

	spaceRequired, _ := strconv.ParseInt(Deployment.ObjectMeta.Annotations["cpu-bandwidth"], 10, 64)

	Node_List, err := Get_MPI_Nodes()

	Wait_For_Nodes_Resources_Update()
	for _, Node := range Node_List.Items {
		node_cpu_limit_temporary[Node.ObjectMeta.Name] = Node_Cpu_Limit[Node.ObjectMeta.Name]
		node_has_mpi[Node.ObjectMeta.Name] = "NO"
		node_has_data_center[Node.ObjectMeta.Name] = "NO"
	}

	for it := 0; it < int(Replicas); it++ {

		replica_fits = 0
		for _, Node := range Node_List.Items {
			if Check_If_Node_Is_Excluded(&Node) == true {
				continue
			}

			if Check_If_Node_Is_Included(&Node) == false {
				continue
			}

			if Allow_MPI_Colocation == 0 {
				if node_has_mpi[Node.ObjectMeta.Name] == "YES" {
					continue
				}
				if Mpi_Job_Exists_Check(Node) == 1 {
					node_has_mpi[Node.ObjectMeta.Name] = "YES"
					continue
				}
			} else if Allow_MPI_Colocation == 1 {
				if node_has_mpi[Node.ObjectMeta.Name] == "YES" {
					continue
				}
			}

			if Scheduling_Policy == 0 {
				if Data_Center_Job_Exists_Check(Node) == 1 {
					continue
				}
			}

			var allocatableCores int
			allocatableCores = int(Node_Capacity_Of_Node(Node.ObjectMeta.Name))
			freeSpace := (allocatableCores - int(node_cpu_limit_temporary[Node.ObjectMeta.Name]))

			if int64(freeSpace) >= spaceRequired {
				node_cpu_limit_temporary[Node.ObjectMeta.Name] = node_cpu_limit_temporary[Node.ObjectMeta.Name] + int64(spaceRequired)
				node_has_mpi[Node.ObjectMeta.Name] = "YES"
				replica_fits = 1
				break
			}
		}
		if replica_fits == 0 {
			if Deployment.ObjectMeta.Annotations["scheduled"] == "off" {
				fmt.Println("MPI JOB worker failed to fit at any Node")
			}
			return 0
		}
	}
	fmt.Println("mpi scheduled = on")
	Deployment.ObjectMeta.Annotations["scheduled"] = "on"
	Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})
	return 1
}

//Check_If_Data_Center_Service_Is_Scheduled checks if any Data-Center Job is on Placement phase
func Check_If_Data_Center_Service_Is_Scheduled() int32 {
	Pods, _ := Get_Pods_all_Namespaces()
	for it := 0; it < len(Pods.Items); it++ {
		Deployment, _ := Get_Deployment_Of_Pod(&Pods.Items[it])
		if Deployment.Status.ReadyReplicas != Deployment.Status.Replicas && Deployment.ObjectMeta.Annotations["scheduled"] == "on" {
			fmt.Println("Another Deployment is being scheduled Right now !")
			return 1
		}
	}
	return 0
}

// Datacenter_Service_Check_If_It_Fits iterates over the Clusters Nodes in order to check if a DATACENTER service deployment can fit onto the Cluster.
// If enough resources are available it returns 1, in any other cases returns 0
// Used before the deployment of each Service in order to check if it fits
func Datacenter_Service_Check_If_It_Fits(Pod *v1.Pod) int32 {

	var replica_fits int
	var node_cpu_limit_temporary = make(map[string]int64)
	var Replicas int32

	if Check_If_MPI_Job_Is_Scheduled() == 1 || Check_If_Data_Center_Service_Is_Scheduled() == 1 {
		return 0
	}

	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}

	Deployments := Client_Set.AppsV1().Deployments(Pod.ObjectMeta.Namespace)
	Deployment, err := Get_Deployment_Of_Pod(Pod)
	if err != nil {
		return 0
	}
	if Deployment.ObjectMeta.Annotations["type"] != "DATACENTER-JOB" {
		return 0
	}
	Replicas = Deployment.Status.Replicas

	fmt.Println("Checking for ", Replicas, " Replicas")

	spaceRequired, _ := strconv.ParseInt(Deployment.ObjectMeta.Annotations["cpu-bandwidth"], 10, 64)

	Node_List, err := Get_All_Nodes()

	Wait_For_Nodes_Resources_Update()
	for _, Node := range Node_List.Items {
		node_cpu_limit_temporary[Node.ObjectMeta.Name] = Node_Cpu_Limit[Node.ObjectMeta.Name]
	}

	for it := 0; it < int(Replicas); it++ {

		replica_fits = 0
		for _, Node := range Node_List.Items {
			if Check_If_Node_Is_Excluded(&Node) == true {
				continue
			}

			if Check_If_Node_Is_Included(&Node) == false {
				continue
			}

			if Scheduling_Policy == 0 {
				if Mpi_Job_Exists_Check(Node) == 1 {
					continue
				}
			}

			var allocatableCores int

			allocatableCores = int(Node_Capacity_Of_Node(Node.ObjectMeta.Name))
			freeSpace := (allocatableCores - int(node_cpu_limit_temporary[Node.ObjectMeta.Name]))

			if int64(freeSpace) >= spaceRequired {
				node_cpu_limit_temporary[Node.ObjectMeta.Name] = node_cpu_limit_temporary[Node.ObjectMeta.Name] + int64(spaceRequired)
				replica_fits = 1
				break
			}
		}
		if replica_fits == 0 {
			if Deployment.ObjectMeta.Annotations["scheduled"] == "off" {
				fmt.Println("Datacenter service failed to fit at any Node")
			}
			return 0
		}
	}
	fmt.Println("Datacenter Pod scheduled = on")
	Deployment.ObjectMeta.Annotations["scheduled"] = "on"
	Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})
	return 1
}
