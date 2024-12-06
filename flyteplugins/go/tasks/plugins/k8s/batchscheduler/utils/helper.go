package utils

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func UpdateLabels(wanted map[string]string, objectMeta *metav1.ObjectMeta) {
	for key, value := range wanted {
		if _, exist := objectMeta.Labels[key]; !exist {
			objectMeta.Labels[key] = value
		}
	}
}

func UpdateAnnotations(wanted map[string]string, objectMeta *metav1.ObjectMeta) {
	for key, value := range wanted {
		if _, exist := objectMeta.Annotations[key]; !exist {
			objectMeta.Annotations[key] = value
		}
	}
}

func UpdatePodTemplateAnnotatations(wanted map[string]string, pod *v1.PodTemplateSpec) {
	UpdateAnnotations(wanted, &pod.ObjectMeta)
}

func UpdatePodTemplateLabels(wanted map[string]string, pod *v1.PodTemplateSpec) {
	UpdateLabels(wanted, &pod.ObjectMeta)
}

type Group struct {
	Master *v1.PodTemplateSpec
	ExtraMasterResource *v1.ResourceList
	Workers []*v1.PodTemplateSpec
	WorkerReplica []int32
}

func NewGroup() Group {
	return Group{
		Master: nil,
		ExtraMasterResource: nil,
		Workers: make([]*v1.PodTemplateSpec, 0),
		WorkerReplica: make([]int32, 0),
	}
}
