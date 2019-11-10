package main

import (
	"context"
	"flag"
	"strings"
	"time"

	"github.com/toventang/debezium-client/adapter"
	"github.com/toventang/debezium-client/client"
	"github.com/toventang/debezium-client/subscriber"
)

func main() {
	var (
		kafkaAddress, groupID, topics                              string
		dstType, dstAddress, dstDatabase, dstUsername, dstPassword string
		timeout                                                    int
	)

	flag.StringVar(&kafkaAddress, "KAFKA_ADDRESS", "192.168.50.199:9092", "kafka addresses")
	flag.StringVar(&groupID, "KAFKA_GROUPID", "cdc.catalogs.subscriber", "group id")
	flag.StringVar(&topics, "KAFKA_TOPICS", "catalogdbs.public.catalogs,catalogdbs.public.templates", "topics")

	flag.StringVar(&dstType, "DST_TYPE", "postgres", "destination database type, support only 'elasticsearch' now")
	flag.StringVar(&dstAddress, "DST_ADDRESS", "192.168.50.199:5432", "destination database addresses")
	flag.StringVar(&dstDatabase, "DST_DATABASE", "postgres", "database name")
	flag.IntVar(&timeout, "DST_TIMEOUT", 5, "R/W timeout")
	flag.StringVar(&dstUsername, "DST_USER", "ecs", "user auth")
	flag.StringVar(&dstPassword, "DST_PASSWORD", "123456", "user auth")
	flag.Parse()

	var tables []string
	t := strings.Split(topics, ",")
	for _, tn := range t {
		s := strings.SplitAfterN(tn, ".", 2)
		tables = append(tables, s[1])
	}

	ctx := context.Background()
	opts := client.Options{
		SubscriberOptions: subscriber.Options{
			Addresses: strings.Split(kafkaAddress, ","),
			GroupID:   groupID,
			Topics:    strings.Split(topics, ","),
		},
		AdapterOptions: adapter.Options{
			ConnectorType: adapter.ParseConnectorType(dstType),
			Addresses:     strings.Split(dstAddress, ","),
			Database:      dstDatabase,
			Timeout:       time.Duration(timeout) * time.Second,
			Tables:        tables,
			Username:      dstUsername,
			Password:      dstPassword,
		},
	}
	cli, err := client.NewClient(opts)
	if err != nil {
		panic(err)
	}

	if err := cli.Start(ctx); err != nil {
		panic(err)
	}
	defer cli.Close()
}
