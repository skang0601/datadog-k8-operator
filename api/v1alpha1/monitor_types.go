/*
Copyright 2020 Sung Kang.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package v1alpha1

import (
	"github.com/zorkian/go-datadog-api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	MonitorFinalizerName = "monitor.datadog.finalizer.skang0601.github.io"
)

// MonitorSpec defines the desired state of Monitor
type MonitorSpec struct {
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Query   string  `json:"query"`
	Message string  `json:"message,omitempty"`
	Options Options `json:"options,omitempty"`

	// Optional list of tags to attach to the monitor.
	// +optional
	Tags []string `json:"tags,omitempty"`
}

// MonitorStatus defines the observed state of Monitor
type MonitorStatus struct {
	Active      bool         `json:"active,omitempty"`
	Id          int32        `json:"id,omitempty"`
	Url         string       `json:"url,omitempty"`
	Error       string       `json:"error,omitempty"`
	LastUpdated *metav1.Time `json:"last_updated,omitempty"`
	Created     *metav1.Time `json:"created,omitempty"`
}

// +kubebuilder:object:root=true

// Monitor is the Schema for the monitors API
type Monitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MonitorSpec   `json:"spec,omitempty"`
	Status MonitorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MonitorList contains a list of Monitor
type MonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Monitor `json:"items"`
}

/*
* Is there a way to avoid re-declaring these structs and having a conversion function when interacting wtih the API?
 */

type Options struct {
	NoDataTimeframe   int            `json:"no_data_time_frame,omitempty"`
	NotifyAudit       bool           `json:"notify_audit,omitempty"`
	NotifyNoData      bool           `json:"notify_no_data,omitempty"`
	RenotifyInterval  int            `json:"renotify_interval,omitempty"`
	NewHostDelay      int            `json:"new_host_delay,omitempty"`
	EvaluationDelay   int            `json:"evaluation_delay,omitempty"`
	Silenced          map[string]int `json:"silenced,omitempty"`
	TimeoutH          int            `json:"timeout_h,omitempty"`
	EscalationMessage string         `json:"escalation_message,omitempty"`

	// +nullable
	Thresholds        Thresholds  `json:"thresholds,omitempty"`
	IncludeTags       bool        `json:"include_tags,omitempty"`
	RequireFullWindow bool        `json:"requre_full_window,omitempty"`
	Locked            bool        `json:"locked,omitempty"`
	EnableLogsSample  bool        `json:"enable_logs_sample,omitempty"`
	QueryConfig       QueryConfig `json:"query_config,omitempty"`
}

type Thresholds struct {
	// +nullable
	Ok string `json:"ok,omitempty"`
	// +nullable
	Critical string `json:"critical,omitempty"`
	// +nullable
	Warning string `json:"warning,omitempty"`
	// +nullable
	Unknown string `json:"unknown,omitempty"`
	// +nullable
	CriticalRecovery string `json:"critical_recovery,omitempty"`
	// +nullable
	WarningRecovery string `json:"warning_recovery,omitempty"`
	// +nullable
	Period Period `json:"period,omitempty"`
	// +nullable
	TimeAggregator string `json:"timeAggregator,omitempty"`
}

type Period struct {
	Seconds int32  `json:"seconds,omitempty"`
	Text    string `json:"text,omitempty"`
	Value   string `json:"value,omitempty"`
	Name    string `json:"name,omitempty"`
	Unit    string `json:"unit,omitempty"`
}

type QueryConfig struct {
	LogSet        LogSet    `json:"logset,omitempty"`
	TimeRange     TimeRange `json:"timeRange,omitempty"`
	QueryString   string    `json:"queryString,omitempty"`
	QueryIsFailed bool      `json:"queryIsFailed,omitempty"`
}

type LogSet struct {
	ID   int32  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type TimeRange struct {
	To   int32 `json:"to,omitempty"`
	From int32 `json:"from,omitempty"`
	Live bool  `json:"live,omitempty"`
}

// Helper function to manage Finalizers in this resource
func (m *Monitor) IsSubmitted() bool {
	if m.Status.Active == false || m.Status.Url == "" {
		return false
	}
	return true
}

func (m *Monitor) IsBeingDeleted() bool {
	return !m.ObjectMeta.DeletionTimestamp.IsZero()
}

func (m *Monitor) HasFinalizer(finalizerName string) bool {
	return containsString(m.ObjectMeta.Finalizers, finalizerName)
}

func (m *Monitor) AddFinalizer() {
	m.ObjectMeta.Finalizers = append(m.ObjectMeta.Finalizers, MonitorFinalizerName)
}

func (m *Monitor) RemoveFinalizer() {
	m.ObjectMeta.Finalizers = removeString(m.ObjectMeta.Finalizers, MonitorFinalizerName)
}

// Helper function to convert internal CRD type to the Monitor type found in the SDK
// Currently we support the minimal set of configurations of the Monitor.
// Most of the future work will be expanding this set and validation of the inputs.
func (m *Monitor) ToApi() *datadog.Monitor {
	var monitor datadog.Monitor

	monitor = datadog.Monitor{
		Type:    &m.Spec.Type,
		Query:   &m.Spec.Query,
		Name:    &m.Spec.Name,
		Message: &m.Spec.Message,
		Tags:    m.Spec.Tags,
		Options: m.Spec.Options.toApi(),
	}
	return &monitor
}

func (o *Options) toApi() *datadog.Options {

	return &datadog.Options{
		NoDataTimeframe:   datadog.NoDataTimeframe(o.NoDataTimeframe),
		NotifyAudit:       &o.NotifyAudit,
		NotifyNoData:      &o.NotifyNoData,
		RenotifyInterval:  &o.RenotifyInterval,
		NewHostDelay:      &o.NewHostDelay,
		EvaluationDelay:   &o.EvaluationDelay,
		Silenced:          o.Silenced,
		TimeoutH:          &o.TimeoutH,
		EscalationMessage: &o.EscalationMessage,
		Thresholds:        o.Thresholds.toApi(),
		ThresholdWindows:  nil,
		IncludeTags:       &o.IncludeTags,
		RequireFullWindow: &o.RequireFullWindow,
		Locked:            &o.Locked,
		EnableLogsSample:  &o.EnableLogsSample,
		QueryConfig:       nil,
	}
}

func (q *QueryConfig) toApi() *datadog.QueryConfig {
	queryConfig := datadog.QueryConfig{
		LogSet:        q.LogSet.toApi(),
		TimeRange:     nil,
		QueryString:   nil,
		QueryIsFailed: nil,
	}

	return &queryConfig
}

func (l *LogSet) toApi() *datadog.LogSet {
	logSet := datadog.LogSet{
		ID:   nil,
		Name: nil,
	}
	return &logSet
}

func (t *Thresholds) toApi() *datadog.ThresholdCount {
	thresholds := datadog.ThresholdCount{
		Ok:               toJsonNumber(t.Ok),
		Critical:         toJsonNumber(t.Critical),
		Warning:          toJsonNumber(t.Warning),
		Unknown:          toJsonNumber(t.Unknown),
		CriticalRecovery: toJsonNumber(t.CriticalRecovery),
		WarningRecovery:  toJsonNumber(t.WarningRecovery),
		Period:           nil,
		TimeAggregator:   nil,
	}

	return &thresholds
}

func (p *Period) toApi() *datadog.Period {
	return &datadog.Period{}
}

func init() {
	SchemeBuilder.Register(&Monitor{}, &MonitorList{})
}
