// Copyright 2023 The distributed-mariadb-controller Authors
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

package sakura

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/sakura-internet/distributed-mariadb-controller/pkg/controller"
)

var (
	// DBControllerStateGaugeVec is the gauge-vec metric in prometheus
	// that holds the current state of the controller.
	DBControllerStateGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "edb_db_controller_state",
			Help: "the controller state of db-controller",
		},
		[]string{"state"},
	)
	// DBControllerStateTransitionCounterVec is the counter-vec metric in prometheus
	// that holds the transition count of the controller.
	DBControllerStateTransitionCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "edb_db_controller_state_transition_count",
			Help: "the counter of the controller state transition",
		},
		[]string{"state"},
	)
)

func init() {
	DBControllerStateGaugeVec.WithLabelValues(string(controller.StateInitial)).Set(1)
	DBControllerStateGaugeVec.WithLabelValues(string(controller.StateFault)).Set(0)
	DBControllerStateGaugeVec.WithLabelValues(string(SAKURAControllerStateCandidate)).Set(0)
	DBControllerStateGaugeVec.WithLabelValues(string(controller.StateReplica)).Set(0)
	DBControllerStateGaugeVec.WithLabelValues(string(controller.StatePrimary)).Set(0)
}

func NewPrometheusMetricRegistry() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		// Go runtime metric collector
		collectors.NewGoCollector(),
		// process metric collector
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),

		// db-controller
		DBControllerStateGaugeVec,
		DBControllerStateTransitionCounterVec,
	)
	return reg
}
