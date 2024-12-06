package kueue

import (
	"fmt"
)

const (
	Kueue = "kueue"
	// labels in the pod template
	Queue         = "kueue.x-k8s.io/queue-name"
	PodGroupName = "kueue.x-k8s.io/pod-group-name"
	PriorityClassName = "kueue.x-k8s.io/priority-class"
	// annotations in the pod template
	PodGroupTotal = "kueue.x-k8s.io/pod-group-total-count"
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
	if queueName := fmt.Sprintf("%s.%s", project, domain); len(configuredQueue) > 0 {
		ExtraLabels[Queue] = fmt.Sprintf("%s.%s", queueName, configuredQueue)
	} else {
		ExtraLabels[Queue] = fmt.Sprintf("%s.%s", queueName, jobNamespace)
	}
}

func cleanLabels() {
	ExtraLabels = make(map[string]string, 0)
	ExtraAnnotations = make(map[string]string, 0)
}

