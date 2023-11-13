// Package model represents metric model calculation helpers.
package model

// Model represents metric model calculation helpers.
type Model struct {
	Services                    int // total count of services
	Nodes                       int // total count of nodes
	ServicesPerNode             int // total count of services per node
	MetricPerService            int // total count of metrics per service
	MetricCardinalityPerService int // total unique count of metric attribute set per service
	MetricPerNode               int // total count of metrics per node
	Metrics                     int // instant unique count of metrics
	DeploysPerMinute            int // total count of deploys per minute
	TotalCardinality            int // instant unique count of metric attribute set
}

// TotalResources returns total count of unique resources.
func (m Model) TotalResources() int {
	return m.Nodes * m.ServicesPerNode
}
