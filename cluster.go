package main

type Cluster struct {
	baseDir string
	conf    ClusterConf
}

func NewCluster(baseDir string, conf ClusterConf) Cluster {
	return Cluster{
		baseDir: baseDir,
		conf:    conf,
	}
}

func (cluster Cluster) CreateNodes() {
	panic("Not implemented")
}

func (cluster Cluster) Start() {
	panic("Not implemented")
}

func (cluster Cluster) Stop() {
	panic("Not implemented")
}

func (cluster Cluster) Cli() {
	panic("Not implemented")
}
