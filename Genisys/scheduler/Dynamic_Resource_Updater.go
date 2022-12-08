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
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
)

var pidc *PIDController = NewPIDController(1, 0, 0)   //add p,i,d to Config file
var pidnet *PIDController = NewPIDController(1, 0, 0) //add p,i,d to Config file
var pidio *PIDController = NewPIDController(1, 0, 0)

type Request struct {
	Items []val `json:"items"`
}

type val struct {
	Value string `json:"value"`
}

// Get_Pod_Metric() takes as input the Pod name of a spesified Pod
// and uses the REST API in order to fetch the Pods Performance Metric from the Prometheus Metric server.
// finally returns the Performance Metric
func Get_Pod_Metric(podname string, pod_namespace string) (metricvalue float64, err error) {

	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}
	Data, err := Client_Set.RESTClient().Get().AbsPath("apis/custom.metrics.k8s.io/v1beta1/namespaces/" + pod_namespace + "/pods/" + podname + "/custom_metric").DoRaw(context.Background())
	if err != nil {
		return 0, err
	}

	s := Request{}
	err1 := json.Unmarshal(Data, &s)
	if err1 != nil {
		panic(err1)
	}
	Value_Final, _ := strconv.Atoi(s.Items[0].Value)
	if strings.HasSuffix(s.Items[0].Value, "m") {
		Value := strings.TrimSuffix(s.Items[0].Value, "m")
		Value_Final, err = strconv.Atoi(Value)
		Value_Final = Value_Final / 1000
		if err != nil {
			return 0, err
		}
	}

	if strings.HasSuffix(s.Items[0].Value, "k") {
		Value := strings.TrimSuffix(s.Items[0].Value, "k")
		Value_Final, err = strconv.Atoi(Value)
		Value_Final = Value_Final * 1000
		if err != nil {
			return 0, err
		}
	}

	if strings.HasSuffix(s.Items[0].Value, "M") {
		Value := strings.TrimSuffix(s.Items[0].Value, "M")
		Value_Final, err = strconv.Atoi(Value)
		Value_Final = Value_Final * 1000000
		if err != nil {
			return 0, err
		}
	}

	f := float64(Value_Final)
	return f, nil
}

func Float_To_Str(input_num float64) string {
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

// PI_Transformer() takes as input a Performance Metric
// and scales it to a static sized target in order to tune the PID controllers
// It returns the relative Scale between the Target and Current Performance Metric
func PI_Transformer(PI float64) float64 {
	var Scale float64
	if PI != 0 {
		Scale = PID_Scale_Final / PI
	} else {
		Scale = 0
	}
	return Scale
}

//Wait_Pod() sleeps for one second, used in when polling Prometheus for a Performance Metric
func Wait_Pod() {
	time.Sleep(1000 * time.Millisecond)
}

const Shell_To_Use = "/bin/sh"

var Some_Map_Mutex = sync.RWMutex{}

//Shellout() takes as input a bash command and executes it to the Host
func Shellout(command string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	if err := exec.CommandContext(ctx, Shell_To_Use, "-c", command).Run(); err != nil {

	}
}

var Services_Names = make(map[string]string)
var Services_Min_Performance = make(map[string]float64)
var Pid_Scaler_Const = make(map[string]float64)
var Lower_Scale_Range = make(map[string]float64)
var Upper_Scale_Range = make(map[string]float64)
var Pid_Net_Scaler_Const = make(map[string]float64)

// Polling_Metrics() iterates over the available Data-Center deployments and scales their resources according the Target Performance Metric.
// As a 1st step it finds the suitable Pod from a given Deployment
// As a 2st step it requests the current Performance Metric for the selected Pod from the Prometheus Metrics Server
// As a 3st step it computes the Ratio between the current Performance Metric and the Target Performance Metric
// As a 4st step if the Ratio is between 0.9 and 1.1 then Polling_Metrics() procceeds to the next deployment without changing any resources
// As a 5th step if the Ratio is outside the specified threshold then Polling_Metrics() follows the resource balance proccedure
////It cycles between 3 kind of resources (CPU, NET, IO), in each iteration it modifies on of these resources
////in order to find the suitable resources to allocate in each step, Polling_Metrics()calls a tuned PID controller that takes as input the Current and Target Performance Metrics
////and returns the resource amount to resize the Pod's container

func Polling_Metrics(Scaler float64, Net_Scaler float64) {

	var Wait_Rounds int = 0

	Wait := 1
	//Deployment_Net_Limit := make(map[string]int64)

	Client_Set, err := kubernetes.NewForConfig(Config)
	if err != nil {
		panic(err.Error())
	}

	for _, namespace := range Namespaces {
		Pods, err := Get_Pods_Of_Namespace(namespace)
		if err != nil {
			continue
		}
		for _, podItem := range Pods.Items {
			Deployment, err := Get_Deployment_Of_Pod(&podItem)
			if err != nil {
				continue
			}
			Deployment.Spec.Template.Spec.Containers[0].Resources.Limits = make(map[v1.ResourceName]resource.Quantity)
		}
	}

	pidc.SetOutputLimits(-1000000, 1000000) //cpu pid
	pidnet.SetOutputLimits(-80000, 80000)   //PID_Network_Output pid
	pidio.SetOutputLimits(-100000, 100000)  //io pid

	go func() {
		for {
			select {
			case <-done:
				return
			case _ = <-ticker.C:
				for _, namespace := range Namespaces {
					Deployments := Client_Set.AppsV1().Deployments(namespace)
					Pods, err := Get_Pods_Of_Namespace(namespace)
					if err != nil {
						continue
					}

					for it := 0; it < len(Pods.Items); it++ {

						var PID_Cpu_Output, PID_Network_Output, PID_IO_Output, Current_Performance, Target_Performance, OPI_For_Printing, TPI_For_Printing float64

						Deployment, err := Get_Deployment_Of_Pod(&Pods.Items[it])
						if err != nil {
							continue
						}

						Deployment_Name, err := Get_Deployment_Name_Of_Pod(&Pods.Items[it])
						if err != nil {
							continue
						}

						if Pods.Items[it].Spec.SchedulerName != *Scheduler_Name {
							continue
						}

						if Pods.Items[it].ObjectMeta.Annotations["app"] == "MPI-JOB" || Deployment.ObjectMeta.Annotations["dynamic-resource-management"] != "1" || Pods.Items[it].Status.Phase != "Running" {
							continue
						}

						//Resource_step is responsible for the controller to choose the approprite pid controller PID_Cpu_Output (this can be cpu,memory,PID_Network_Output,disk io)
						Resource_Step, _ := strconv.ParseFloat(Deployment.ObjectMeta.Annotations["resource"], 64)

						Target_Performance, _ = strconv.ParseFloat(Deployment.ObjectMeta.Annotations["metric-value"], 64)

						//Initialize pid for each metric
						pidc.Set(PID_Scale_Final)
						pidnet.Set(PID_Scale_Final)
						pidio.Set(PID_Scale_Final)

						//Wait Pod to respond
						fmt.Printf("Waiting response from Pod : %s\n", Pods.Items[it].ObjectMeta.Name)
						Current_Performance, err = Get_Pod_Metric(Pods.Items[it].ObjectMeta.Name, Pods.Items[it].ObjectMeta.Namespace)
						if err != nil && Wait_Rounds < 10 { // check Pod status if pending then skip
							it--
							Deployments = Client_Set.AppsV1().Deployments(namespace)
							Pods, _ = Get_Pods_Of_Namespace(namespace)

							Wait_Rounds++
							continue
						}
						Wait_Rounds = 0
						Wait = 0

						//Keep metrics unchanged for printing
						OPI_For_Printing = Current_Performance
						TPI_For_Printing = Target_Performance

						Current_Performance = Current_Performance * PI_Transformer(Target_Performance)

						Target_Performance = PID_Scale_Final

						Ratio := Current_Performance / Target_Performance

						//PID controller outputs
						PID_Cpu_Output = pidc.UpdateDuration(Current_Performance, 1*time.Second)
						PID_Network_Output = pidnet.UpdateDuration(Current_Performance, 1*time.Second)
						PID_IO_Output = pidio.UpdateDuration(Current_Performance, 1*time.Second)

						Shellout("kubectl top pods " + Pods.Items[it].ObjectMeta.Name + " | grep 'm' | awk '{print $2}' | sed 's/m//' > real_time_cpu ")
						real_cpu, err := ioutil.ReadFile("real_time_cpu") // just pass the file name
						if err != nil {
							fmt.Print(err)
						}

						real_cpu_str := string(real_cpu) // convert content to a 'string'
						real_cpu_usage, _ := strconv.ParseFloat(strings.TrimSpace(real_cpu_str), 8)

						//here we print all information about the current state of each Pod
						fmt.Println("==========================================================\n")
						fmt.Printf("POD NAME: %s\n\n", Pods.Items[it].ObjectMeta.Name)
						fmt.Printf("POD LOCATION: %s\n\n", Pods.Items[it].Spec.NodeName)
						fmt.Println("CPU ALLOCATION (MILLICORES): " + fmt.Sprintf("%f", float64(Deployment_Cpu_Limit[Deployment_Name+namespace])))
						fmt.Println("CPU USAGE (MILLICORES): " + real_cpu_str)
						fmt.Println("NETWORK BANDWIDTH ALLOCATION (MB/s): " + fmt.Sprintf("%f", float64(Deployment_Net_Limit[Deployment_Name+namespace])))
						fmt.Println("IO BANDWIDTH ALLOCATION (MB/s): " + fmt.Sprintf("%f", float64(Deployment_IO_Limit[Deployment_Name+namespace])/1048576))
						fmt.Printf("PERFORMANCE TARGET VALUE: %f\n", TPI_For_Printing)
						fmt.Printf("CURRENT PERFORMANCE VALUE %f\n", OPI_For_Printing)
						fmt.Printf("RATIO (CURRENT/TARGET) %f\n", Ratio)
						fmt.Printf("RESOURCE STEP (1 CPU, 2 NETWORK, 3 DISK IO) %f\n", Resource_Step)
						fmt.Printf("PID CPU OUTPUT (MILLICORES) %f\n", PID_Cpu_Output)
						fmt.Printf("PID NETWORK OUTPUT (MB/S) %f\n", PID_Network_Output)
						fmt.Printf("PID IO OUTPUT (MB) %f\n", PID_IO_Output)
						fmt.Println("==========================================================\n")

						//write logs in csv format
						err = os.Chmod("plots/"+Deployment_Name+".csv", 0777)
						if err != nil {
							//	continue
						}
						f, err := os.OpenFile("plots/"+Deployment_Name+".csv", os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModeAppend)
						if err != nil {
							//	continue
						}
						defer f.Close()

						//scv column Data
						oPI_Str := fmt.Sprintf("%f", OPI_For_Printing)
						tPI_Str := fmt.Sprintf("%f", TPI_For_Printing)
						Cpu_Str := fmt.Sprintf("%f", float64(Deployment_Cpu_Limit[Deployment_Name+namespace])/(100*12))
						Net_Str := fmt.Sprintf("%f", float64(Deployment_Net_Limit[Deployment_Name+namespace])/(8000*120))
						IO_Str := fmt.Sprintf("%f", float64(Deployment_IO_Limit[Deployment_Name+namespace])/(1048576*120))

						Time_Now := time.Now()
						Elapsed_Time := Time_Now.Sub(Start)
						Time_Secs_Str := strconv.FormatInt(int64(Elapsed_Time/time.Millisecond)/1000, 10)
						Plot_Data["plots/"+Deployment_Name+".csv"] = nil

						if Lines_In_File_Counter("plots/"+Deployment_Name+".csv") == 0 {
							Plot_Data["plots/"+Deployment_Name+".csv"] = append(Plot_Data["plots/"+Deployment_Name+".csv"], []string{"time_x", "current_performance", "target_performance", "cpu_node_percentage", "net_node_percentage", "io_node_percentage"})
						}
						//append Data to csv
						Plot_Data["plots/"+Deployment_Name+".csv"] = append(Plot_Data["plots/"+Deployment_Name+".csv"], []string{Time_Secs_Str, oPI_Str, tPI_Str, Cpu_Str, Net_Str, IO_Str})

						w := csv.NewWriter(f)
						w.WriteAll(Plot_Data["plots/"+Deployment_Name+".csv"])

						if err := w.Error(); err != nil {
							log.Fatal(err)
						}

						//In this section the apropriate changes are made to the Pod resources by the controller
						Wait_For_Nodes_Resources_Update()

						//TODO Reduce the number of nested if statements
						if Ratio < 0.9 {
							if OPI_For_Printing >= 0 && OPI_For_Printing < 1000 {
								if real_cpu_usage < 50 && real_cpu_usage > 0 {
									if *Deployment.Spec.Replicas > 1 {
										Target_Performance_Init, _ := strconv.ParseFloat(Deployment.ObjectMeta.Annotations["metric-value-init"], 64)
										Target_Performance_New := Target_Performance_Init / float64((*Deployment.Spec.Replicas)-1)
										*Deployment.Spec.Replicas = (*Deployment.Spec.Replicas) - 1
										Deployment.ObjectMeta.Annotations["metric-value"] = fmt.Sprintf("%d", int(Target_Performance_New))
									}
								}
								Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})
								continue
							}
							if (Node_Cpu_Limit[Pods.Items[it].Spec.NodeName] + int64(Scaler*PID_Cpu_Output)) < Total_Node_Capacity {
								Deployment.ObjectMeta.Annotations["cpu-bandwidth"] = fmt.Sprintf("%d", Deployment_Cpu_Limit[Deployment_Name+namespace]+int64(Scaler*PID_Cpu_Output))

							} else if (Deployment_Cpu_Limit[Deployment_Name+namespace]+int64(Scaler*PID_Cpu_Output)) > Total_Node_Capacity && Deployment.ObjectMeta.Annotations["hpa-running"] == "off" && Deployment.ObjectMeta.Annotations["HPA"] == "1" {
								Deployment.ObjectMeta.Annotations["cpu-bandwidth"] = fmt.Sprintf("%d", (Deployment_Cpu_Limit[Deployment_Name+namespace]/2 + int64(Scaler*PID_Cpu_Output)/2))
								*Deployment.Spec.Replicas = 2 * (*Deployment.Spec.Replicas)
								Target_Performance_UN, _ := strconv.ParseFloat(Deployment.ObjectMeta.Annotations["metric-value"], 64)
								Deployment.ObjectMeta.Annotations["metric-value"] = fmt.Sprintf("%d", int(Target_Performance_UN/2))
								Deployment.ObjectMeta.Annotations["hpa-running"] = "on"

							} else {
								Deployment.ObjectMeta.Annotations["cpu-bandwidth"] = fmt.Sprintf("%d", Deployment_Cpu_Limit[Deployment_Name+namespace]+int64(Scaler*PID_Cpu_Output))
								Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})
								Shellout("kubectl delete Pod " + Pods.Items[it].ObjectMeta.Name + " -n " + namespace)
								time.Sleep(1000 * time.Millisecond)
							}
							// if Resource_Step == 1 {
							// 	Deployment.ObjectMeta.Annotations["resource"] = strconv.FormatFloat(Resource_Step+1, 'f', 6, 64)
							// 	if (Node_Cpu_Limit[Pods.Items[it].Spec.NodeName] + int64(Scaler*PID_Cpu_Output)) < Total_Node_Capacity {
							// 		Deployment.ObjectMeta.Annotations["cpu-bandwidth"] = fmt.Sprintf("%d", Deployment_Cpu_Limit[Deployment_Name+namespace]+int64(Scaler*PID_Cpu_Output))
							// 	} else {
							// 		Deployment.ObjectMeta.Annotations["cpu-bandwidth"] = fmt.Sprintf("%d", Deployment_Cpu_Limit[Deployment_Name+namespace]+int64(Scaler*PID_Cpu_Output))
							// 		Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})
							// 		Shellout("kubectl delete Pod " + Pods.Items[it].ObjectMeta.Name + " -n " + namespace)
							// 	}
							// } else if Resource_Step == 2 {
							// 	Deployment.ObjectMeta.Annotations["resource"] = strconv.FormatFloat(Resource_Step+1, 'f', 6, 64)
							// 	if (Node_Net_Limit[Pods.Items[it].Spec.NodeName] + int64(Net_Scaler*PID_Network_Output)) < Total_Node_Capacity_Net {
							// 		Deployment.ObjectMeta.Annotations["net-bandwidth"] = fmt.Sprintf("%d", Deployment_Net_Limit[Deployment_Name+namespace]+int64(Net_Scaler*PID_Network_Output))
							// 	} else {
							// 		Deployment.ObjectMeta.Annotations["net-bandwidth"] = fmt.Sprintf("%d", Deployment_Net_Limit[Deployment_Name+namespace]+int64(Net_Scaler*PID_Network_Output))
							// 		Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})
							// 		Shellout("kubectl delete Pod " + Pods.Items[it].ObjectMeta.Name + " -n " + namespace)
							// 	}
							// } else if Resource_Step == 3 {
							// 	Deployment.ObjectMeta.Annotations["resource"] = strconv.FormatFloat(1, 'f', 6, 64)
							// 	if (Node_IO_Limit[Pods.Items[it].Spec.NodeName] + Deployment_IO_Limit[Deployment_Name+namespace] + int64(180*PID_IO_Output)) < 300000000 {
							// 		Deployment.ObjectMeta.Annotations["io-bandwidth"] = fmt.Sprintf("%d", Deployment_IO_Limit[Deployment_Name+namespace]+int64(30*PID_IO_Output))
							// 	} else {
							// 		Deployment.ObjectMeta.Annotations["io-bandwidth"] = fmt.Sprintf("%d", Deployment_IO_Limit[Deployment_Name+namespace]+int64(30*PID_IO_Output))
							// 	}
							// }
							Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})
							if Resource_Step < 2 {
								Wait_Pod()
							}

						} else if Ratio > 1.2 {
							if (Node_Cpu_Limit[Pods.Items[it].Spec.NodeName] + int64(Scaler*PID_Cpu_Output)) < Total_Node_Capacity {
								if Deployment_Cpu_Limit[Deployment_Name+namespace]+int64(Scaler*PID_Cpu_Output) > 500 {
									Deployment.ObjectMeta.Annotations["cpu-bandwidth"] = fmt.Sprintf("%d", Deployment_Cpu_Limit[Deployment_Name+namespace]+int64(Scaler*PID_Cpu_Output))
								} else {
									Deployment.ObjectMeta.Annotations["cpu-bandwidth"] = fmt.Sprintf("%d", 500)
								}
							}
							// if Resource_Step == 1 {
							// 	Deployment.ObjectMeta.Annotations["resource"] = strconv.FormatFloat(Resource_Step+1, 'f', 6, 64) // experimental low resource migration bug fix
							// 	if (Node_Cpu_Limit[Pods.Items[it].Spec.NodeName] + int64(Scaler*PID_Cpu_Output)) < Total_Node_Capacity {
							// 		if Deployment_Cpu_Limit[Deployment_Name+namespace]+int64(Scaler*PID_Cpu_Output) > 0 {
							// 			Deployment.ObjectMeta.Annotations["cpu-bandwidth"] = fmt.Sprintf("%d", Deployment_Cpu_Limit[Deployment_Name+namespace]+int64(Scaler*PID_Cpu_Output))
							// 		} else {
							// 			Deployment.ObjectMeta.Annotations["cpu-bandwidth"] = fmt.Sprintf("%d", 500)
							// 		}
							// 	}
							// } else if Resource_Step == 2 {
							// 	Deployment.ObjectMeta.Annotations["resource"] = strconv.FormatFloat(Resource_Step+1, 'f', 6, 64)
							// 	if (Node_Net_Limit[Pods.Items[it].Spec.NodeName] + int64(Net_Scaler*PID_Network_Output)) < Total_Node_Capacity_Net {
							// 		if Deployment_Net_Limit[Deployment_Name+namespace]+int64(Net_Scaler*PID_Network_Output) > 8000 {
							// 			Deployment.ObjectMeta.Annotations["net-bandwidth"] = fmt.Sprintf("%d", Deployment_Net_Limit[Deployment_Name+namespace]+int64(Net_Scaler*PID_Network_Output))
							// 		}
							// 	} else {
							// 		Deployment.ObjectMeta.Annotations["net-bandwidth"] = fmt.Sprintf("%d", Deployment_Net_Limit[Deployment_Name+namespace]+int64(Net_Scaler*PID_Network_Output))
							// 	}

							// } else if Resource_Step == 3 {
							// 	Deployment.ObjectMeta.Annotations["resource"] = strconv.FormatFloat(1, 'f', 6, 64)
							// 	if (Node_IO_Limit[Pods.Items[it].Spec.NodeName] + int64(70*PID_IO_Output)) < 300000000 {
							// 		Deployment.ObjectMeta.Annotations["io-bandwidth"] = fmt.Sprintf("%d", Deployment_IO_Limit[Deployment_Name+namespace]+int64(15*PID_IO_Output))
							// 	} else {
							// 		Deployment.ObjectMeta.Annotations["io-bandwidth"] = fmt.Sprintf("%d", Deployment_IO_Limit[Deployment_Name+namespace]+int64(15*PID_IO_Output))
							// 	}
							// }

							Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})
							if Resource_Step < 2 {
								Wait_Pod()
							}

						} else {
							Deployment.ObjectMeta.Annotations["resource"] = strconv.FormatFloat(1, 'f', 6, 64)
							//if real_cpu_usage > 800 {
							//	Deployment.ObjectMeta.Annotations["cpu-bandwidth"] = fmt.Sprintf("%d", int64(real_cpu_usage))
							//}
							//Deployments.Update(context.Background(), Deployment, metav1.UpdateOptions{})
						}

						Net_Limit, _ := strconv.ParseFloat(Deployment.ObjectMeta.Annotations["net-bandwidth"], 64)
						IO_Limit, _ := strconv.ParseFloat(Deployment.ObjectMeta.Annotations["io-bandwidth"], 64)

						Cpu_Limit := Deployment_Cpu_Limit[Deployment_Name+namespace]
						Net_Limit = Net_Limit / 8000
						Net_Limit_String := fmt.Sprintf("%f", Net_Limit)
						IO_Limit = IO_Limit / 1000000
						Io_Limit_String := fmt.Sprintf("%f", IO_Limit)

						//here the controller writes logs for each Pod to the path monitoring/

						if _, err := os.Stat("monitoring/"); os.IsNotExist(err) {
							err := os.Mkdir("monitoring", 0777)
							if err != nil {
								panic(err)
							}
						}

						if Current_Performance > 10 {
							f, err := os.OpenFile("monitoring/"+Deployment_Name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0660)
							check(err)
							w := bufio.NewWriter(f)
							_, err = w.WriteString(" CPU limit : " + fmt.Sprintf("%d", Cpu_Limit) + " NETWORK limit : " + Net_Limit_String + " IO limit : " + Io_Limit_String + " oPI : " + fmt.Sprintf("%f", Current_Performance) + " tPI : " + fmt.Sprintf("%f", Target_Performance) + " " + time.Now().Format(time.RFC850) + " \n")
							check(err)
							w.Flush()

							Node_List_Log, err := os.OpenFile("monitoring/"+Deployment_Name+"_node_list", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0660)
							check(err)
							w2 := bufio.NewWriter(Node_List_Log)
							_, err = w2.WriteString(Pods.Items[it].Spec.NodeName + " " + time.Now().Format(time.RFC850) + " \n")
							check(err)
							w2.Flush()
						}
						time.Sleep(1000 * time.Duration(time.Millisecond))
					}
					time.Sleep(1000 * time.Millisecond) //avoid spinning
				}
			}
		}
	}()
}

func Resource_Controller() {

	go Polling_Metrics(0.015, 2)

}
