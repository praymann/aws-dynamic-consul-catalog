package rds

import (
	"github.com/spf13/cobra"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	observer "github.com/imkira/go-observer"
	cache "github.com/patrickmn/go-cache"
	cc "github.com/seatgeek/aws-dynamic-consul-catalog/backend/consul"
	"github.com/seatgeek/aws-dynamic-consul-catalog/config"
	gelf "github.com/seatgeek/logrus-gelf-formatter"
	log "github.com/sirupsen/logrus"
)

// RDS ...
type RDS struct {
	rds              *rds.RDS
	backend          config.Backend
	logger           log.Entry
	instanceFilters  config.Filters
	tagFilters       config.Filters
	tagCache         *cache.Cache
	checkInterval    time.Duration
	quitCh           chan int
	onDuplicate      string
	servicePrefix    string
	serviceSuffix    string
	consulNodeName   string
	consulMasterTag  string
	consulReplicaTag string
}

func New(c *cobra.Command) *RDS {

	logLevel, err := log.ParseLevel(strings.ToUpper(c.Flags().Lookup("log-level").Value.String()))
	if err != nil {
		log.Fatalf("%s (%s)", err, c.Flags().Lookup("log-level").Value.String())
	}
	log.SetLevel(logLevel)

	logFormat := strings.ToLower(c.Flags().Lookup("log-format").Value.String())
	switch logFormat {
	case "json":
		log.SetFormatter(new(gelf.GelfFormatter))
	case "text":
		log.SetFormatter(new(log.TextFormatter))
	default:
		log.Fatalf("log-format value %s is not a valid option (json or text)", logFormat)
	}

	sess := session.Must(session.NewSession())

	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(sess),
			},
		})

	tagCacheDuration, _ := c.Flags().GetDuration("rds-tag-cache-time")
	checkInterval, _ := c.Flags().GetDuration("check-interval")
	instanceFilters, _ := c.Flags().GetStringSlice("instance-filter")
	tagFilters, _ := c.Flags().GetStringSlice("tag-filter")

	return &RDS{
		rds: rds.New(session.Must(session.NewSession(&aws.Config{
			Credentials: creds,
		}))),
		backend:          cc.NewBackend(),
		instanceFilters:  config.ProcessFilters(instanceFilters),
		tagFilters:       config.ProcessFilters(tagFilters),
		tagCache:         cache.New(tagCacheDuration, 30*time.Minute),
		checkInterval:    checkInterval,
		quitCh:           make(chan int),
		onDuplicate:      c.Flags().Lookup("on-duplicate").Value.String(),
		servicePrefix:    c.Flags().Lookup("consul-service-prefix").Value.String(),
		serviceSuffix:    c.Flags().Lookup("consul-service-suffix").Value.String(),
		consulNodeName:   c.Flags().Lookup("consul-node-name").Value.String(),
		consulMasterTag:  c.Flags().Lookup("consul-leader-tag").Value.String(),
		consulReplicaTag: c.Flags().Lookup("consul-follower-tag").Value.String(),
	}
}

func (r *RDS) Run() {
	log.Info("Starting RDS app")

	allInstances := observer.NewProperty(nil)
	filteredInstances := observer.NewProperty(nil)
	catalogState := &config.CatalogState{}

	go r.backend.CatalogReader(catalogState, r.consulNodeName, r.quitCh)
	go r.reader(allInstances)
	go r.filter(allInstances, filteredInstances)
	go r.writer(filteredInstances, catalogState)

	<-r.quitCh
}
