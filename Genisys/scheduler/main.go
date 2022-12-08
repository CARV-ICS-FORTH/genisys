// Copyright 2016 Google Inc. All Rights Reserved.
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
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
)

var Total_Node_Capacity_Net int64 = 120 * 8000
var Max_Node_Allocation_Percentage float64
var Total_Node_Capacity int64
var CPU_Node_Capacity_Float float64
var PID_Scale_Float float64 = 5
var PID_Scale_Final float64
var Kubernetes_Native int64 = 0

var Timeout_Current_Global int = 100000

var Deployment_Net_Limit = make(map[string]int64)
var Node_Net_Limit = make(map[string]int64)

var Deployment_IO_Limit = make(map[string]int64)
var Node_IO_Limit = make(map[string]int64)

var Deployment_Cpu_Limit = make(map[string]int64)
var Node_Cpu_Limit = make(map[string]int64)

// Scheduling_Policy
// ||DEFINE 0 -> DEDICATED||--||DEFINE 1 -> MIXED||
// DEDICATED -> EACH MPI CONTAINER RUNNING AT DEDICATED NODE TO AVOID INTERFERENCE
// MIXED -> SCHEDULER MANAGES MPI JOBS IN THE SAME WAY AS IT DOES WITH DATACENTER APPLICATIONS
var Scheduling_Policy = 1

var Allow_MPI_Colocation = 0

// Spread_Placement_Policy
// ||DEFINE 0 -> choose the most loaded node||--||DEFINE 1 -> choose least loaded node||
var Spread_Placement_Policy = 0

var Start time.Time

var Kube_Config *string
var Config *rest.Config

var Global_Namespace = ""

var Scheduler_Name *string
var Kube_Config_Path *string
var Excluded_Nodes []string
var Included_Nodes []string
var Namespaces []string

var Timer_Node_Updater int
var Timer_Deployment_Updater int

var ticker *time.Ticker
var done chan bool

func WaitForCtrlC() {
	var end_waiter sync.WaitGroup
	end_waiter.Add(1)
	var signal_channel chan os.Signal
	signal_channel = make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	go func() {
		<-signal_channel
		end_waiter.Done()
	}()
	end_waiter.Wait()
}

func main() {

	ticker = time.NewTicker(500 * time.Millisecond)
	done = make(chan bool)

	log.Println("Starting custom scheduler...")
	log.Println("Press Ctrl+C to exit the scheduler")
	/* Parse the CLI arguements*/
	//Nodes to exclude from cluster
	Exclude_Nodelist := flag.String("Exclude_Nodelist", "", "-Exclude_Nodelist <Node-1,Node-2..,Node-n>")

	//Nodes to include from cluster
	Include_Nodelist := flag.String("Include_Nodelist", "", "-Include_Nodelist <Node-1,Node-2..,Node-n>")

	//Max Node Capacity
	Max_Node_Capacity := flag.String("Max_Node_Capacity", "1", "-Max_Node_Capacity 1 (100% Node Capacity)")

	Scheduling_Policy_Input := flag.String("Allow_Task_Colocation", "1", "-Allow_Task_Colocation 1 (0 disable Colocation)")

	//Enable Spreading Placement
	Spread_Placement_Policy_Input := flag.String("Allow_Spreading_of_Tasks", "1", "-Allow_Spreading 1 (0 disable spreading across the least loaded nodes)")

	//Allow MPI Job colocation over the same nodes
	Allow_MPI_Colocation_Input := flag.String("Allow_MPI_Colocation", "1", "-Allow_MPI_Colocation 1 (0 disbale MPI JOB Colocation)")

	//PID Scaling Const
	PID_Scale := flag.String("PID_Scale", "3", "-PID_Scale 5 (PID Scaler larger values accelerate resource scaling while losing accuracy)")

	//Namespaces to include
	Namespaces_Input := flag.String("Namespaces", "default", "-Namespaces <namespace-1,namespace-2..,namespace-n>")

	//Scheduler name DEFAULT "genisys"
	Scheduler_Name = flag.String("Scheduler_Name", "genisys", "-Scheduler Name (DEFAULT = genisys)")

	//Path to the Kubernetes Config file
	Kube_Config_Path = flag.String("Kube_Config_Path", "/home/master/.kube/Config", "Path to Kubernetes Configuration file")

	flag.Parse()

	//Fetch Kubernetes Config
	Kube_Config = flag.String("Kube_Config", *Kube_Config_Path, "absolute path to the Kube_Config file")
	var err error
	Config, err = clientcmd.BuildConfigFromFlags("", *Kube_Config)
	if err != nil {
		Config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	PID_Scale_Final, _ = strconv.ParseFloat(*PID_Scale, 64)
	PID_Scale_Final = PID_Scale_Final * 10000

	Max_Node_Allocation_Percentage, _ = strconv.ParseFloat(*Max_Node_Capacity, 64)

	// Fetch the CPU capacity from the nodes
	var CPU_Node_Capacity string
	CPU_Node_Capacity, _ = Node_Capacity()

	CPU_Node_Capacity_Float, _ = strconv.ParseFloat(CPU_Node_Capacity, 64)

	Total_Node_Capacity = int64(CPU_Node_Capacity_Float * 1000 * Max_Node_Allocation_Percentage)

	Scheduling_Policy, _ = strconv.Atoi(*Scheduling_Policy_Input)
	Spread_Placement_Policy, _ = strconv.Atoi(*Spread_Placement_Policy_Input)

	Allow_MPI_Colocation, _ = strconv.Atoi(*Allow_MPI_Colocation_Input)

	Excluded_Nodes = strings.Split(*Exclude_Nodelist, ",")
	Included_Nodes = strings.Split(*Include_Nodelist, ",")

	Namespaces = strings.Split(*Namespaces_Input, ",")

	Node_List, err := Get_All_Nodes()

	for _, Node := range Node_List.Items {
		if Check_If_Node_Is_Excluded(&Node) == true {
			continue
		}

		if Check_If_Node_Is_Included(&Node) == false {
			continue
		}
		Shellout("kubectl label node " + Node.ObjectMeta.Name + " mpi=1")
	}

	//log Start scheduling timestamp
	Start = time.Now()
	doneChan := make(chan struct{})

	Per_App_Log_File_Reset()
	Per_App_Log_File_Creator()
	go Resources_Bandwidth_Updater_Deployments()
	go Resource_Bandwidth_Updater_Nodes()
	go Resources_Bandwidth_Printer_Nodes()
	go Resources_Cpu_Limiter_Virtual_Clusters()
	go Pod_Scheduler()
	go Slurm_Job_Deleter()

	//Start resource controller
	Resource_Controller()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	WaitForCtrlC()
	ticker.Stop()
	done <- true

	for {
		select {
		case <-signalChan:
			log.Printf("Shutdown signal received, exiting...")
			close(doneChan)
			os.Exit(0)
		}
	}
}
