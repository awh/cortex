package chunk

import (
	"time"

	"github.com/prometheus/common/model"
)

// Parse a time and return it, or panic.
func mustParse(isoTimestamp string) model.Time {
	t, err := time.Parse(time.RFC3339, isoTimestamp)
	if err != nil {
		panic(err)
	}
	return model.TimeFromUnix(t.UnixNano())
}

// AliasLookup is a made up name for something which can find
// a list of known alias for a metric name, scoped to a tenant.
type AliasLookup interface {
	// Return known aliases for the specified metric. Guarantees:
	// Aliases are returned in ascending order of start time
	// Each alias overlaps with the preceeding and subsequent aliases only
	// The first alias has `from` of `model.Earliest`
	// The last alias has an `until` of `model.Latest`
	GetAliases(orgID, metricName string) []MetricNameAlias
}

type MetricNameAlias struct {
	from  model.Time
	until model.Time
	name  string
}

type nilAliasLookup struct {
}

func (nal *nilAliasLookup) GetAliases(orgID, metricName string) []MetricNameAlias {
	return nil
}

type exampleAliasLookup struct {
}

func (eal *exampleAliasLookup) GetAliases(orgID, metricName string) []MetricNameAlias {
	if metricName == "node_cpu_seconds_total" {
		// We (ignoring tenant for the purposes of example) have rolled out a
		// new Prometheus node exporter on 2019-04-16 that renames node_cpu ->
		// node_cpu_seconds_total. Prior to 12:00 all nodes are running the old
		// exporter, and after 12:05 they are all running the new one; during
		// that five minute rollout window we are scraping a mixture of both
		// across nodes.
		return []MetricNameAlias{
			MetricNameAlias{model.Earliest, mustParse("2019-04-16T12:00:05"), "node_cpu"},
			MetricNameAlias{mustParse("2019-04-16T12:00:00"), model.Latest, "node_cpu_seconds_total"},
		}
	} else {
		// No known aliases.
		return nil
	}
}
