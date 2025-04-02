// Copyright 2025 The distributed-mariadb-controller Authors
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

package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	// dbControllerStateGaugeVec is the gauge-vec metric in prometheus
	// that holds the current state of the controller.
	dbControllerStateGaugeVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "edb_db_controller_state",
			Help: "the controller state of db-controller",
		},
		[]string{"state"},
	)
	// dbControllerStateTransitionCounterVec is the counter-vec metric in prometheus
	// that holds the transition count of the controller.
	dbControllerStateTransitionCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "edb_db_controller_state_transition_count",
			Help: "the counter of the controller state transition",
		},
		[]string{"state"},
	)
)

func init() {
	dbControllerStateGaugeVec.WithLabelValues(string(StateInitial)).Set(1)
	dbControllerStateGaugeVec.WithLabelValues(string(StateFault)).Set(0)
	dbControllerStateGaugeVec.WithLabelValues(string(StateCandidate)).Set(0)
	dbControllerStateGaugeVec.WithLabelValues(string(StateReplica)).Set(0)
	dbControllerStateGaugeVec.WithLabelValues(string(StatePrimary)).Set(0)
}

func NewPrometheusMetricRegistry() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		// Go runtime metric collector
		collectors.NewGoCollector(),
		// process metric collector
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),

		// db-controller
		dbControllerStateGaugeVec,
		dbControllerStateTransitionCounterVec,
	)
	return reg
}
