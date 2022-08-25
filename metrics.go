package main

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	hostMetrics      = make(map[string]*prometheus.GaugeVec)
	clusterMetrics   = make(map[string]*prometheus.GaugeVec)
	vMMetrics        = make(map[string]*prometheus.GaugeVec)
	datastoreMetrics = make(map[string]*prometheus.GaugeVec)
	clusterIdToName  = make(map[int]string) //name of the cluster, as name not given in vm api information
)

func initMetrics() {

	/*
		Host Metrics setup
	*/

	{

		// Memory

		hostMetrics["TotalMem"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_host_totalmem",
			Help: "total memory available on host",
		}, []string{"cluster", "host"})

		hostMetrics["MaxMem"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_host_maxmem",
			Help: "maximum memory available on host (total - reserved)",
		}, []string{"cluster", "host"})

		hostMetrics["MemUsage"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_host_memusage",
			Help: "total allocated memory on host",
		}, []string{"cluster", "host"})

		//CPU

		hostMetrics["TotalCPU"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_host_totalcpu",
			Help: "total cpu available on host",
		}, []string{"cluster", "host"})

		hostMetrics["MaxCPU"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_host_maxcpu",
			Help: "maximum cpu available on host (total - reserved)",
		}, []string{"cluster", "host"})

		hostMetrics["CPUUsage"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_host_cpuusage",
			Help: "total allocated cpu on host",
		}, []string{"cluster", "host"})

		// Nr of VMs

		hostMetrics["RunningVMs"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_host_runningvms",
			Help: "running virtual machines on host",
		}, []string{"cluster", "host"})

		hostMetrics["State"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_host_state",
			Help: "state of host",
		}, []string{"cluster", "host", "state"})
	}

	/*
		Cluster Metrics setup
	*/

	{

		//Memory
		clusterMetrics["TotalMem"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_cluster_totalmem",
			Help: "total memory available in cluster",
		}, []string{"cluster"})

		clusterMetrics["MaxMem"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_cluster_maxmem",
			Help: "maximum memory available in cluster (total - reserved)",
		}, []string{"cluster"})

		clusterMetrics["MemUsage"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_cluster_memusage",
			Help: "total allocated memory in cluster",
		}, []string{"cluster"})

		//CPU

		clusterMetrics["TotalCPU"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_cluster_totalcpu",
			Help: "total cpu available in cluster",
		}, []string{"cluster"})

		clusterMetrics["MaxCPU"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_cluster_maxcpu",
			Help: "maximum cpu available in cluster (total - reserved",
		}, []string{"cluster"})

		clusterMetrics["CPUUsage"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_cluster_cpuusage",
			Help: "total allocated cpu in cluster",
		}, []string{"cluster"})
	}

	/*
		VM Metrics setup
	*/

	{
		vMMetrics["State"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_vm_state",
			Help: "state of the vm",
		}, []string{"cluster", "host", "vm", "state"})

		vMMetrics["LCMState"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_vm_lcmstate",
			Help: "lcm state of the vm. Only set when vm is active (State 3)",
		}, []string{"cluster", "host", "vm", "lcm_state"})
	}
	/*
		Datastore Metrics setup
	*/
	{
		datastoreMetrics["Available"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_datastore_available",
			Help: "which datastores are available in which cluster",
		}, []string{"cluster", "datastore"})

		datastoreMetrics["State"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_datastore_state",
			Help: "datastore state (ready, disabled)",
		}, []string{"datastore", "state"})

		datastoreMetrics["TotalMB"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_datastore_totalmb",
			Help: "total capacity of datastore",
		}, []string{"datastore"})

		datastoreMetrics["UsedMB"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_datastore_usedmb",
			Help: "used capacity of datastore",
		}, []string{"datastore"})

		datastoreMetrics["FreeMB"] = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "one_datastore_freemb",
			Help: "free capacity of datastore",
		}, []string{"datastore"})
	}
}

func recordMetrics(config config, logger log.Logger) {

	level.Info(logger).Log("msg", "recording metrics from opennebula frontend", "interval", config.interval)

	client := goca.NewDefaultClient(goca.NewConfig(config.user, config.password, config.endpoint))
	controller := goca.NewController(client)

	for {
		fillIdMaps(controller, logger)
		recordHostMetrics(controller, logger)
		recordVmMetrics(controller, logger)
		recordDataStoreMetrics(controller, logger)

		time.Sleep(time.Duration(config.interval) * time.Second)
	}
}

func fillIdMaps(controller *goca.Controller, logger log.Logger) {
	clusters, err := controller.Clusters().Info(true)
	if err != nil {
		level.Error(logger).Log("msg", "error retrieving cluster info", "error", err)
		return
	}
	for _, cluster := range clusters.Clusters {
		clusterIdToName[cluster.ID] = cluster.Name
	}
}

func recordHostMetrics(controller *goca.Controller, logger log.Logger) {

	pool, err := controller.Hosts().Info()
	if err != nil {
		level.Error(logger).Log("msg", "error retrieving hosts info", "error", err)
		return
	}

	type metrics struct {
		cluster, metric string
	}
	sum := make(map[metrics]int)

	//reset the state metrics, as the State name is stored as label
	hostMetrics["State"].Reset()
	for _, host := range pool.Hosts {

		state, err := host.State()
		if err != nil {
			level.Error(logger).Log("msg", "error retrieving state from host", "host", host.Name, "error", err)
			continue
		}
		level.Debug(logger).Log("msg", "host metrics",
			"host", host.Name,
			"state", state,
			"TotalMem", host.Share.TotalMem,
			"MaxMem", host.Share.MaxMem,
			"MemUsage", host.Share.MemUsage,
			"TotalCPU", host.Share.TotalCPU,
			"CPUUsage", host.Share.CPUUsage,
			"RunningVMs", host.Share.RunningVMs,
		)

		// record host metrics
		hostMetrics["TotalMem"].With(prometheus.Labels{"cluster": host.Cluster, "host": host.Name}).Set(float64(host.Share.TotalMem))
		hostMetrics["MaxMem"].With(prometheus.Labels{"cluster": host.Cluster, "host": host.Name}).Set(float64(host.Share.MaxMem))
		hostMetrics["MemUsage"].With(prometheus.Labels{"cluster": host.Cluster, "host": host.Name}).Set(float64(host.Share.MemUsage))
		hostMetrics["TotalCPU"].With(prometheus.Labels{"cluster": host.Cluster, "host": host.Name}).Set(float64(host.Share.TotalCPU))
		hostMetrics["MaxCPU"].With(prometheus.Labels{"cluster": host.Cluster, "host": host.Name}).Set(float64(host.Share.MaxCPU))
		hostMetrics["CPUUsage"].With(prometheus.Labels{"cluster": host.Cluster, "host": host.Name}).Set(float64(host.Share.CPUUsage))
		hostMetrics["RunningVMs"].With(prometheus.Labels{"cluster": host.Cluster, "host": host.Name}).Set(float64(host.Share.RunningVMs))
		hostMetrics["State"].With(prometheus.Labels{"cluster": host.Cluster, "host": host.Name, "state": state.String()}).Set(float64(host.StateRaw))

		// sum cluster metrics
		sum[metrics{host.Cluster, "TotalMem"}] = sum[metrics{host.Cluster, "TotalMem"}] + host.Share.TotalMem
		sum[metrics{host.Cluster, "MaxMem"}] = sum[metrics{host.Cluster, "MaxMem"}] + host.Share.MaxMem
		sum[metrics{host.Cluster, "MemUsage"}] = sum[metrics{host.Cluster, "MemUsage"}] + host.Share.MemUsage
		sum[metrics{host.Cluster, "TotalCPU"}] = sum[metrics{host.Cluster, "TotalCPU"}] + host.Share.TotalCPU
		sum[metrics{host.Cluster, "MaxCPU"}] = sum[metrics{host.Cluster, "MaxCPU"}] + host.Share.MaxCPU
		sum[metrics{host.Cluster, "CPUUsage"}] = sum[metrics{host.Cluster, "CPUUsage"}] + host.Share.CPUUsage
	}

	for key, value := range sum {
		// record cluster metrics
		clusterMetrics[key.metric].With(prometheus.Labels{"cluster": key.cluster}).Set(float64(value))
	}
}

func recordVmMetrics(controller *goca.Controller, logger log.Logger) {
	vms, err := controller.VMs().Info()
	if err != nil {
		level.Error(logger).Log("msg", "error retrieving vms info", "error", err)
		return
	}

	//reset the state metrics, as the State name is stored as label
	vMMetrics["State"].Reset()
	vMMetrics["LCMState"].Reset()
	for _, vm := range vms.VMs {

		state, lcmState, err := vm.State()
		if err != nil {
			level.Error(logger).Log("msg", "error retrieving state from vm", "vm", vm.Name, "error", err)
			continue
		}

		if len(vm.HistoryRecords) == 0 {
			continue
		}

		host := vm.HistoryRecords[0].Hostname
		cluster := clusterIdToName[vm.HistoryRecords[0].CID]

		level.Debug(logger).Log("msg", "vm metrics",
			"vm", vm.Name,
			"StateRaw", vm.StateRaw,
			"State", state,
			"LCMStateRaw", vm.LCMStateRaw,
			"LCMState", lcmState,
			"Host", host,
			"Cluster", cluster,
		)

		vMMetrics["State"].With(prometheus.Labels{"cluster": cluster, "host": host, "vm": vm.Name, "state": state.String()}).Set(float64(vm.StateRaw))
		vMMetrics["LCMState"].With(prometheus.Labels{"cluster": cluster, "host": host, "vm": vm.Name, "lcm_state": lcmState.String()}).Set(float64(vm.LCMStateRaw))
	}
}

func recordDataStoreMetrics(controller *goca.Controller, logger log.Logger) {

	datastores, err := controller.Datastores().Info()
	if err != nil {
		level.Error(logger).Log("msg", "error retrieving datastores info", "error", err)
		return
	}

	//reset the state metrics, as the State name is stored as label
	datastoreMetrics["State"].Reset()
	for _, datastore := range datastores.Datastores {

		state, err := datastore.State()
		if err != nil {
			level.Debug(logger).Log("msg", "Failed get state from datastore", "error", err)
			continue
		}

		level.Debug(logger).Log("msg", "datastore metrics",
			"datastore", datastore.Name,
			"StateRaw", datastore.StateRaw,
			"State", state,
			"TotalMB", datastore.TotalMB,
			"UsedMB", datastore.UsedMB,
			"FreeMB", datastore.FreeMB,
			"DiskType", datastore.DiskType,
		)

		for _, clusterId := range datastore.Clusters.ID {
			cluster := clusterIdToName[clusterId]
			datastoreMetrics["Available"].With(prometheus.Labels{"datastore": datastore.Name, "cluster": cluster}).Set(float64(1))
		}

		datastoreMetrics["TotalMB"].With(prometheus.Labels{"datastore": datastore.Name}).Set(float64(datastore.TotalMB))
		datastoreMetrics["UsedMB"].With(prometheus.Labels{"datastore": datastore.Name}).Set(float64(datastore.UsedMB))
		datastoreMetrics["FreeMB"].With(prometheus.Labels{"datastore": datastore.Name}).Set(float64(datastore.FreeMB))
		datastoreMetrics["State"].With(prometheus.Labels{"datastore": datastore.Name, "state": state.String()}).Set(float64(state))
	}

}
