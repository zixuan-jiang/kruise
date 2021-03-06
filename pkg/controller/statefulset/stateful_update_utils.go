/*
Copyright 2019 The Kruise Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package statefulset

import (
	appsv1alpha1 "github.com/openkruise/kruise/pkg/apis/apps/v1alpha1"
	"github.com/openkruise/kruise/pkg/util/updatesort"
	v1 "k8s.io/api/core/v1"
)

func sortPodsToUpdate(rollingUpdateStrategy *appsv1alpha1.RollingUpdateStatefulSetStrategy, updateRevision string, replicas []*v1.Pod) []int {
	var updateMin int
	var priorityStrategy *appsv1alpha1.UpdatePriorityStrategy
	if rollingUpdateStrategy != nil {
		if rollingUpdateStrategy.Partition != nil {
			updateMin = int(*rollingUpdateStrategy.Partition)
		}
		if rollingUpdateStrategy.UnorderedUpdate != nil {
			priorityStrategy = rollingUpdateStrategy.UnorderedUpdate.PriorityStrategy
		}
	}
	maxUpdate := len(replicas) - updateMin
	if maxUpdate <= 0 {
		return []int{}
	}

	var updatedIdxs []int
	var waitUpdateIdxs []int
	for target := len(replicas) - 1; target >= 0; target-- {
		if isTerminating(replicas[target]) {
			updatedIdxs = append(updatedIdxs, target)
		} else if getPodRevision(replicas[target]) == updateRevision {
			updatedIdxs = append(updatedIdxs, target)
		} else {
			waitUpdateIdxs = append(waitUpdateIdxs, target)
		}
	}

	if priorityStrategy != nil {
		waitUpdateIdxs = updatesort.NewPrioritySorter(priorityStrategy).Sort(replicas, waitUpdateIdxs)
	}

	allIdxs := append(updatedIdxs, waitUpdateIdxs...)
	if len(allIdxs) > maxUpdate {
		allIdxs = allIdxs[:maxUpdate]
	}

	return allIdxs
}
