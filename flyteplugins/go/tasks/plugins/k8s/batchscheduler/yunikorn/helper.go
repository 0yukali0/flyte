package yunikorn

import (
	"encoding/json"
	"fmt"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/utils"
)

const (
	// scheduler name in the pod spec
	Yunikorn = "yunikorn"
	// Labels in the pod template
	AppID = "applicationId"
	Queue = "queue"
	// Annotations in the pod template
	TaskGroupNameKey    = "yunikorn.apache.org/task-group-name"
	TaskGroupsKey       = "yunikorn.apache.org/task-groups"
	TaskGroupParameters = "yunikorn.apache.org/schedulingPolicyParameters"
	PreemptionAllow     = "yunikorn.apache.org/allow-preemption"
)

var (
	ExtraLabels      = make(map[string]string, 0)
	ExtraAnnotations = make(map[string]string, 0)
)

func GetCommonLabels() (map[string]string, map[string]string) {
	defer cleanLabels()
	return ExtraLabels, ExtraAnnotations
}

func CreateCommLabels(name string, project string, domain string, jobNamespace string, configuredQueue string) {
	ExtraLabels[AppID] = GenerateTaskGroupAppID(name)
	if queueName := fmt.Sprintf("root.%s.%s", project, domain); len(configuredQueue) > 0 {
		ExtraLabels[Queue] = fmt.Sprintf("%s.%s", queueName, configuredQueue)
	} else {
		ExtraLabels[Queue] = fmt.Sprintf("%s.%s", queueName, jobNamespace)
	}
}

func cleanLabels() {
	ExtraLabels = make(map[string]string, 0)
	ExtraAnnotations = make(map[string]string, 0)
}

func MutateRayJob(app *rayv1.RayJob, gangschedulingParameter string) error {
	g := utils.Group{}
	rayjobSpec := &app.Spec
	appSpec := rayjobSpec.RayClusterSpec
	g.Master = &appSpec.HeadGroupSpec.Template
	for _, worker := range appSpec.WorkerGroupSpecs {
		g.Workers = append(g.Workers, &worker.Template)
		g.WorkerReplica = append(g.WorkerReplica, *worker.Replicas)
	}
	if ok := *appSpec.EnableInTreeAutoscaling; ok {
		res := v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("500m"),
			v1.ResourceMemory: resource.MustParse("512Mi"),
		}
		g.ExtraMasterResource = &res
	}
	return UpdateTaskGroupAnnotations(g, gangschedulingParameter)
}

func UpdateTaskGroupAnnotations(target utils.Group, gangschedulingParameter string) error {
	TaskGroups := make([]TaskGroup, 1)
	for index, worker := range target.Workers {
		worker.Spec.SchedulerName = YUNIKORN
		meta := worker.ObjectMeta
		spec := worker.Spec
		name := GetTaskGroupSlaveName(index)
		TaskGroups = append(TaskGroups, TaskGroup{
			Name:                      name,
			MinMember:                 target.WorkerReplica[index],
			Labels:                    meta.Labels,
			Annotations:               meta.Annotations,
			MinResource:               Allocation(spec.Containers),
			NodeSelector:              spec.NodeSelector,
			Affinity:                  spec.Affinity,
			TopologySpreadConstraints: spec.TopologySpreadConstraints,
		})
		meta.Annotations[TaskGroupNameKey] = name
		meta.Annotations[PreemptionAllow] = "false"
	}
	master := target.Master
	meta := master.ObjectMeta
	spec := master.Spec
	spec.SchedulerName = YUNIKORN
	res := Allocation(spec.Containers)
	if target.ExtraMasterResource != nil {
		res = Add(res, *target.ExtraMasterResource)
	}
	headname := GetTaskGroupMasterName()
	TaskGroups[0] = TaskGroup{
		Name:                      headname,
		MinMember:                 1,
		Labels:                    meta.Labels,
		Annotations:               meta.Annotations,
		MinResource:               res,
		NodeSelector:              spec.NodeSelector,
		Affinity:                  spec.Affinity,
		TopologySpreadConstraints: spec.TopologySpreadConstraints,
	}
	meta.Annotations[TaskGroupNameKey] = headname
	info, err := json.Marshal(TaskGroups)
	if err != nil {
		return err
	}
	meta.Annotations[TaskGroupsKey] = string(info[:])
	meta.Annotations[PreemptionAllow] = "false"
	if len(gangschedulingParameter) > 0 {
		meta.Annotations[TaskGroupParameters] = gangschedulingParameter
	}
	return nil
}
