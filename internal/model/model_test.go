package model

import "testing"

func TestModel(t *testing.T) {
	m := Model{
		Services:                    10_000,
		Nodes:                       100_000,
		ServicesPerNode:             100,
		DeploysPerMinute:            10,
		MetricPerNode:               100,
		MetricCardinalityPerService: 100,
		MetricPerService:            25,
		Metrics:                     10_000,
		TotalCardinality:            1_000_000_000, // 1B
	}

	t.Logf("TotalResources: %d", m.TotalResources())
}
