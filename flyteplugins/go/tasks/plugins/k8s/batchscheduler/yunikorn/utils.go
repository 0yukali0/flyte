package yunikorn

import (
	v1 "k8s.io/api/core/v1"
)

func Allocation(containers []v1.Container) v1.ResourceList {
	totalResources := v1.ResourceList{}
	for _, c := range containers {
		for name, q := range c.Resources.Limits {
			if _, exists := totalResources[name]; !exists {
				totalResources[name] = q.DeepCopy()
				continue
			}
			total := totalResources[name]
			total.Add(q)
			totalResources[name] = total
		}
	}
	return totalResources
}

func Add(left v1.ResourceList, right v1.ResourceList) v1.ResourceList {
	result := left
	for name, value := range left {
		sum := value
		if value2, ok := right[name]; ok {
			sum.Add(value2)
			result[name] = sum
		} else {
			result[name] = value
		}
	}
	for name, value := range right {
		if _, ok := left[name]; !ok {
			result[name] = value
		}
	}
	return result
}
