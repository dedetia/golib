package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"time"
)

type (
	Options struct {
		URI            string
		DB             string
		AppName        string
		ConnectTimeout time.Duration
		PingTimeout    time.Duration
	}

	Mongo struct {
		DB *mongo.Database
	}
)

func (opt *Options) init() {
	if opt.AppName == "" {
		opt.AppName = "Default"
	}

	if opt.ConnectTimeout == 0 {
		opt.ConnectTimeout = 10 * time.Second
	}

	if opt.PingTimeout == 0 {
		opt.PingTimeout = 2 * time.Second
	}
}

func New(opt *Options) *Mongo {
	opt.init()

	client, err := connect(opt)
	if err != nil {
		panic(err)
	}
	return &Mongo{
		DB: client.Database(opt.DB),
	}
}

func connect(opt *Options) (*mongo.Client, error) {
	connectCtx, cancelConnectCtx := context.WithTimeout(context.Background(), opt.ConnectTimeout)
	defer cancelConnectCtx()

	cmdMon := otelmongo.NewMonitor()
	opts := []*options.ClientOptions{
		options.Client().SetConnectTimeout(opt.ConnectTimeout).ApplyURI(opt.URI).SetAppName(opt.AppName),
		options.Client().SetMonitor(cmdMon),
	}

	client, err := mongo.Connect(connectCtx, opts...)
	if err != nil {
		return nil, err
	}

	pingCtx, cancelPingCtx := context.WithTimeout(context.Background(), opt.PingTimeout)
	defer cancelPingCtx()

	if err = client.Ping(pingCtx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}
