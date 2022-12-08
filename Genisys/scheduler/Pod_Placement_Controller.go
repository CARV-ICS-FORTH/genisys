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
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kubernetes "k8s.io/client-go/kubernetes"
)

// Check_If_Node_Is_Excluded() checks if a Node is included in the excluded Node list,
// used in order to avoid placing jobs in such Nodes

func Check_If_Node_Is_Excluded(Node *v1.Node) bool {
	for _, node_ex := range Excluded_Nodes {
		if Node.Name == node_ex {
			return true
		}
	}
	return false
}

// Check_If_Node_Is_Included() checks if a Node is included in the included Node list,
// used in order to place jobs in selected Nodes

func Check_If_Node_Is_Included(Node *v1.Node) bool {
	if len(Included_Nodes) == 1 && Included_Nodes[0] == "" {
		return true
	}
	for _, node_in := range Included_Nodes {
		if Node.Name == node_in {
			return true
		}
	}
	return false
}

// Pod_To_Node_Finder() takes as input a Pod and checks if a Node with enough free resources exists in order to place the Pod
// Returns a list with suitable Nodes or nil if no such a Node exists

func Pod_To_Node_Finder(Pod *v1.Pod) ([]v1.Node, error) {

	// In order to find Pod requested resources, we have to fetch the Deployment that it belongs to
	Deployment, err := Get_Deployment_Of_Pod(Pod)
	if err != nil {
		return nil, err
	}
	Deployment_Name, err := Get_Deployment_Name_Of_Pod(Pod)
	if err != nil {
		return nil, err
	}
	Node_List, err := Get_All_Nodes()
	// If the Deployment is MPI type then get only the MPI nodes
	if Deployment.ObjectMeta.Annotations["type"] == "MPI-JOB" {
		Node_List, err = Get_MPI_Nodes()
		if err != nil {
			return nil, err
		}
	}

	Some_Map_Mutex.Lock()
	// Fetch the Deployment's resources in order to schedule the Pod
	if Deployment.ObjectMeta.Annotations["type"] != "MPI-JOB" {
		Deployment_Net_Limit[Deployment_Name+Pod.ObjectMeta.Namespace], _ = strconv.ParseInt(Deployment.ObjectMeta.Annotations["net-bandwidth"], 10, 64)
		Deployment_Net_Limit[Deployment_Name+Pod.ObjectMeta.Namespace] = Deployment_Net_Limit[Deployment_Name+Pod.ObjectMeta.Namespace] * 8000
		Deployment_IO_Limit[Deployment_Name+Pod.ObjectMeta.Namespace], _ = strconv.ParseInt(Deployment.ObjectMeta.Annotations["io-bandwidth"], 10, 64)
		Deployment_Cpu_Limit[Deployment_Name+Pod.ObjectMeta.Namespace], _ = strconv.ParseInt(Deployment.ObjectMeta.Annotations["cpu-bandwidth"], 10, 64)
	}
	Some_Map_Mutex.Unlock()

	var preferedNode v1.Node

	var Nodes []v1.Node
	fitFailures := make([]string, 0)

	var spaceRequired int

	Cores, _ := strconv.ParseInt(Deployment.ObjectMeta.Annotations["cpu-bandwidth"], 10, 64)

	spaceRequired += int(Cores)

	// If the Least Loaded Policy is enabled then prefer the Least Loaded Node
	if Spread_Placement_Policy == 1 {
		for _, Node := range Node_List.Items {
			if Node_Cpu_Limit[Node.ObjectMeta.Name] >= Node_Cpu_Limit[preferedNode.ObjectMeta.Name] {
				preferedNode = Node
			}
		}
	}

	fmt.Println("\n\n\n\n===== Deployment: ", Deployment_Name, " =====")
	for _, Node := range Node_List.Items {

		// Check if the Node is inclued or excluded
		if Check_If_Node_Is_Excluded(&Node) == true {
			continue
		}
		if Check_If_Node_Is_Included(&Node) == false {
			continue
		}

		var allocatableCores int

		// Before scheduling wait for resource update for the Nodes
		Wait_For_Nodes_Resources_Update()
		fmt.Println("===== ", Node.ObjectMeta.Name, " =====")
		allocatableCores = int(Node_Capacity_Of_Node(Node.ObjectMeta.Name))
		fmt.Println("allocatableCores: ", int(float64(allocatableCores)))
		freeSpace := (allocatableCores - int(Node_Cpu_Limit[Node.ObjectMeta.Name]))
		fmt.Println("freeSpace: ", freeSpace)
		fmt.Println("spaceRequired: ", spaceRequired)
		fmt.Println("= = = = = = = = = =")

		// If the Pod does not fit to the Node then advance to the next Node
		if freeSpace < spaceRequired {
			m := fmt.Sprintf("fit failure on Node (%s): Insufficient CPU\n", Node.ObjectMeta.Name)
			fitFailures = append(fitFailures, m)
			continue

		} else if Spread_Placement_Policy == 0 { // If the Most Loaded Policy is enabled the chooes the Most Loaded Node

			if (Node_Net_Limit[Node.ObjectMeta.Name]+Deployment_Net_Limit[Deployment_Name+Pod.ObjectMeta.Namespace]) < Total_Node_Capacity_Net &&
				(Node_IO_Limit[Node.ObjectMeta.Name]+Deployment_IO_Limit[Deployment_Name+Pod.ObjectMeta.Namespace]) < 300000000 &&
				(Node_Cpu_Limit[Node.ObjectMeta.Name]+Deployment_Cpu_Limit[Deployment_Name+Pod.ObjectMeta.Namespace]) <= int64(allocatableCores) &&
				(Node_Cpu_Limit[Node.ObjectMeta.Name] >= Node_Cpu_Limit[preferedNode.ObjectMeta.Name]) {

				if Scheduling_Policy == 0 {
					if Mpi_Job_Exists_Check(Node) == 1 {
						continue
					}
					if Deployment.ObjectMeta.Annotations["type"] == "MPI-JOB" && Data_Center_Job_Exists_Check(Node) == 1 {
						continue
					}
				} else {
					if Deployment.ObjectMeta.Annotations["type"] == "MPI-JOB" && Mpi_Job_Exists_Check(Node) == 1 && Allow_MPI_Colocation == 0 {
						continue
					} else if Deployment.ObjectMeta.Annotations["type"] == "MPI-JOB" && Mpi_Job_Exists_Check(Node) == 1 && Allow_MPI_Colocation == 1 {
						if Mpi_Job_Self_Exists_Check(Deployment_Name, Node, Pod.ObjectMeta.Namespace) == 1 {
							continue
						}
					}
				}
				preferedNode = Node
				Nodes = append(Nodes, preferedNode)
			}

		} else if Spread_Placement_Policy == 1 { // If the Least Loaded Policy is enabled the chooes the Least Loaded Node

			if (Node_Net_Limit[Node.ObjectMeta.Name]+Deployment_Net_Limit[Deployment_Name+Pod.ObjectMeta.Namespace]) < Total_Node_Capacity_Net &&
				(Node_IO_Limit[Node.ObjectMeta.Name]+Deployment_IO_Limit[Deployment_Name+Pod.ObjectMeta.Namespace]) < 300000000 &&
				(Node_Cpu_Limit[Node.ObjectMeta.Name]+Deployment_Cpu_Limit[Deployment_Name+Pod.ObjectMeta.Namespace]) <= int64(allocatableCores) &&
				(Node_Cpu_Limit[Node.ObjectMeta.Name] <= Node_Cpu_Limit[preferedNode.ObjectMeta.Name]) {

				if Scheduling_Policy == 0 {
					if Mpi_Job_Exists_Check(Node) == 1 {
						continue
					}
					if Deployment.ObjectMeta.Annotations["type"] == "MPI-JOB" && Data_Center_Job_Exists_Check(Node) == 1 {
						continue
					}
				} else {
					if Deployment.ObjectMeta.Annotations["type"] == "MPI-JOB" && Mpi_Job_Exists_Check(Node) == 1 && Allow_MPI_Colocation == 0 {
						continue
					} else if Deployment.ObjectMeta.Annotations["type"] == "MPI-JOB" && Mpi_Job_Exists_Check(Node) == 1 && Allow_MPI_Colocation == 1 {
						if Mpi_Job_Self_Exists_Check(Deployment_Name, Node, Pod.ObjectMeta.Namespace) == 1 {
							continue
						}
					}
				}
				preferedNode = Node
				Nodes = append(Nodes, preferedNode)
			}
		}
	}

	// If 0 suitable Nodes are found the Pod does not fit to the Cluster
	if len(Nodes) == 0 {
		fmt.Printf("Pod (%s) failed to fit in any Node\n%s", Pod.ObjectMeta.Name, strings.Join(fitFailures, "\n\n"))
		return nil, err
	} else {
		fmt.Printf("Pod (%s) successfuly fit in Cluster\n%s", Pod.ObjectMeta.Name, strings.Join(fitFailures, "\n\n"))
	}
	return Nodes, nil

}

//If a Pod is on the deletion phase, wait for resource deallocation
func Check_If_Pod_Deletion_Exists() {
	Pods, _ := Get_Pods_all_Namespaces()
	for it := 0; it < len(Pods.Items); it++ {
		if Pods.Items[it].DeletionTimestamp != nil {
			it = 0
			time.Sleep(100 * time.Millisecond)
		}
		Pods, _ = Get_Pods_all_Namespaces()
	}
}

// Schedule_Pod() takes as input a Pod, checks if a Node with enough free resources exists to place the Pod
// and finally calls Bind_Pod_To_Node_Controller function in order to place it to suitable Node

func Schedule_Pod(Pod *v1.Pod) error {

	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}

	Deployments := Client_Set.AppsV1().Deployments(Pod.ObjectMeta.Namespace)
	Deployment, err := Get_Deployment_Of_Pod(Pod)

	if err != nil {
		return err
	}

	//Give initial resources to deployment if deynamic resource management is enabled
	//Initialize resources only on the first pod scheduled
	if Deployment.ObjectMeta.Annotations["dynamic-resource-management"] == "1" && Deployment.ObjectMeta.Annotations["first-time-scheduled"] != "0" && Deployment.ObjectMeta.Annotations["type"] != "MPI-JOB" {
		Deployment.ObjectMeta.Annotations["metric-value-init"] = Deployment.ObjectMeta.Annotations["metric-value"]
		Deployment.ObjectMeta.Annotations["cpu-bandwidth"] = "1000"
		Deployment.ObjectMeta.Annotations["net-bandwidth"] = "1"
		Deployment.ObjectMeta.Annotations["io-bandwidth"] = "15000000"
		Deployment.ObjectMeta.Annotations["first-time-scheduled"] = "0"
	}

	if Deployment.ObjectMeta.Annotations["first-time-scheduled"] != "0" && Deployment.ObjectMeta.Annotations["type"] != "MPI-JOB" {

		//Change "first-time-scheduled" entry from 1 to 0 in order ro prevent reinitializing deployment entries
		Deployment.ObjectMeta.Annotations["first-time-scheduled"] = "0"
		Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})
		Deployments = Client_Set.AppsV1().Deployments(Pod.ObjectMeta.Namespace)
		Deployment, err = Get_Deployment_Of_Pod(Pod)

		//Go Client workaround, try until the Deployment.ObjectMeta.Annotations["first-time-scheduled"] is initialized successfuly
		for Deployment.ObjectMeta.Annotations["first-time-scheduled"] == "1" {
			Deployment.ObjectMeta.Annotations["first-time-scheduled"] = "0"
			Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})
			Deployments = Client_Set.AppsV1().Deployments(Pod.ObjectMeta.Namespace)
			Deployment, err = Get_Deployment_Of_Pod(Pod)
		}

		Deployment.ObjectMeta.Annotations["scheduled"] = "off"
		//If the deployement annotations are not initialized by the user then
		//initialize them on the first schedule attempt
		if Deployment.ObjectMeta.Annotations["cpu-bandwidth"] == "" {
			Deployment.ObjectMeta.Annotations["type"] = "DATACENTER-JOB"
			Deployment.ObjectMeta.Annotations["cpu-bandwidth"] = "1000"
			Deployment.ObjectMeta.Annotations["net-bandwidth"] = "1"
			Deployment.ObjectMeta.Annotations["io-bandwidth"] = "15000000"
		}
	}

	if Deployment.ObjectMeta.Annotations["first-time-scheduled"] != "0" && Deployment.ObjectMeta.Annotations["type"] == "MPI-JOB" {
		Deployment.ObjectMeta.Annotations["first-time-scheduled"] = "0"
		Deployment.ObjectMeta.Annotations["scheduled"] = "off"
	}

	//Forward the update to Kubernetes
	Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})

	Check_If_Pod_Deletion_Exists()

	//Check if the Deployment fits in the Cluster under currrent conditions
	if Mpi_Job_Check_If_It_Fits(Pod) == 0 && Deployment.ObjectMeta.Annotations["scheduled"] != "on" && Deployment.ObjectMeta.Annotations["type"] == "MPI-JOB" {
		err := errors.New("")
		return err
	}

	//Check if the Deployment fits in the Cluster under currrent conditions
	if Datacenter_Service_Check_If_It_Fits(Pod) == 0 && Deployment.ObjectMeta.Annotations["scheduled"] != "on" && Deployment.ObjectMeta.Annotations["type"] != "MPI-JOB" {
		err := errors.New("")
		return err
	}

	Nodes, err := Pod_To_Node_Finder(Pod)
	if err != nil {
		return err
	}
	if len(Nodes) == 0 {
		//HPA add as option on job deployment
		if Deployment.ObjectMeta.Annotations["HPA"] == "1" {
			if Deployment.ObjectMeta.Annotations["hpa-running"] == "on" && Deployment.ObjectMeta.Annotations["type"] != "MPI-JOB" {
				Deployment.ObjectMeta.Annotations["hpa-running"] = "off"
				Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})
			}
		}
		err := errors.New("")
		return err
	}
	if Deployment.ObjectMeta.Annotations["type"] != "MPI-JOB" {
		Deployment.ObjectMeta.Annotations["hpa-running"] = "off"
		Deployment.ObjectMeta.Annotations["resource"] = "1"
		Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})
	}

	err = Bind_Pod_To_Node_Controller(Pod, Nodes[len(Nodes)-1])
	if err != nil {
		return err
	}
	Client_Set, _ = kubernetes.NewForConfig(Config)
	Deployments = Client_Set.AppsV1().Deployments(Pod.ObjectMeta.Namespace)
	Deployment, err = Get_Deployment_Of_Pod(Pod)
	if err != nil {
		return err
	}
	Pod.Spec.NodeName = Nodes[len(Nodes)-1].Name
	return nil
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

// return a condition function that indicates whether the given pod is
// currently running
func isPodRunning(podName string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := Get_Pod_By_Name(podName)
		if err != nil {
			return false, err
		}

		if pod == nil {
			return false, err
		}

		if IsPodReady(pod) == false {
			return false, err
		}
		switch pod.Status.Phase {
		case v1.PodRunning:
			return true, nil
		case v1.PodFailed, v1.PodSucceeded:
			return false, err
		}
		return false, nil
	}
}

// Poll up to Timeout_Current seconds for pod to enter running state.
// Returns an error if the pod never enters the running state.
func waitForPodRunning(podName string, Timeout_Current time.Duration) error {
	return wait.PollImmediate(time.Second, Timeout_Current, isPodRunning(podName))
}

//Pod scheduler iterates over the Pending Pods in order to schedule them
func Pod_Scheduler() {
	for {
		defer timeTrack(time.Now(), "factorial")
		select {
		case <-done:
			return
		case _ = <-ticker.C:
			Pods, _ := Get_Pods_all_Namespaces()
			for it := 0; it < len(Pods.Items); it++ {
				if Pods.Items[it].Spec.SchedulerName != *Scheduler_Name {
					continue
				}
				Deployment, err := Get_Deployment_Of_Pod(&Pods.Items[it])
				if err != nil {
					continue
				}

				// If Another MPI deployment is scheduled then skip
				if Check_If_MPI_Job_Is_Scheduled() == 1 && Deployment.ObjectMeta.Annotations["scheduled"] == "off" {
					continue
				}

				// If Another DATA-CENTER deployment is scheduled then skip
				if Check_If_Data_Center_Service_Is_Scheduled() == 1 && Deployment.ObjectMeta.Annotations["scheduled"] == "off" {
					continue
				}

				// Schedule Pod the are in Pending state
				if string(Pods.Items[it].Status.Phase) == "Pending" {
					// Assign the Pod to a Kubernetes Node
					err := Schedule_Pod(&Pods.Items[it])
					if err != nil {
						fmt.Println(err.Error())
					}
					waitForPodRunning(Pods.Items[it].ObjectMeta.Name, 10*time.Second)
					// If the Pod is scheduled successfuly then wait until it ready and running
					if err == nil {
						Timeout_Current := 0
						for Pods.Items[it].Status.Phase != "Running" {
							Client_Set, err := kubernetes.NewForConfig(Config)
							if err != nil {
								panic(err.Error())
							}

							// If we Pod creation exceeds the timeout then skip to the next Pod
							if Timeout_Current > Timeout_Current_Global {
								if Get_Pod_Namespace(Pods.Items[it].ObjectMeta.Name) == "" {
									break
								}
								err := Client_Set.CoreV1().Pods(Get_Pod_Namespace(Pods.Items[it].ObjectMeta.Name)).Delete(context.TODO(), Pods.Items[it].ObjectMeta.Name, metav1.DeleteOptions{})
								if err != nil {
									log.Fatal(err)
								}
								break
							}

							Pod, _ := Get_Pod_By_Name(Pods.Items[it].ObjectMeta.Name)
							// After the Pod has been scheduled, wait for a full Node resource update
							if Pod == nil {
								Wait_For_Nodes_Resources_Update_To_Start()
								Wait_For_Nodes_Resources_Update()
								break
							}

							Pods.Items[it] = *Pod
							Timeout_Current++
							//Sleep for 1 ms to avoid spinning
							time.Sleep(1 * time.Millisecond)
						}
					}
				}
			}
		}
	}
}
