package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/dedetia/golib/storage/redis/options"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"net/url"
	"strings"
	"time"
)

type (
	SentinelOptions struct {
		URI string
		DB  int
		Exp int
	}
)

func (o *SentinelOptions) failoverOptions() (*redis.FailoverOptions, error) {
	var opts redis.FailoverOptions

	uri, err := url.Parse(o.URI)
	if err != nil {
		return nil, err
	}

	opts.SentinelAddrs = strings.Split(uri.Host, ",")
	if len(opts.SentinelAddrs) < 1 {
		return nil, errors.New("invalid address")
	}

	opts.Password, _ = uri.User.Password()

	opts.MasterName = uri.Query().Get("master")
	if opts.MasterName == "" {
		opts.MasterName = "mymaster"
	}

	q := queryOptions{q: uri.Query()}
	opts.DB = q.int("db")

	opts.Protocol = q.int("protocol")
	opts.ClientName = q.string("client_name")
	opts.MaxRetries = q.int("max_retries")
	opts.MinRetryBackoff = q.duration("min_retry_backoff")
	opts.MaxRetryBackoff = q.duration("max_retry_backoff")
	opts.DialTimeout = q.duration("dial_timeout")
	opts.ReadTimeout = q.duration("read_timeout")
	opts.WriteTimeout = q.duration("write_timeout")
	opts.PoolFIFO = q.bool("pool_fifo")
	opts.PoolSize = q.int("pool_size")
	opts.PoolTimeout = q.duration("pool_timeout")
	opts.MinIdleConns = q.int("min_idle_conns")
	opts.MaxIdleConns = q.int("max_idle_conns")
	opts.MaxActiveConns = q.int("max_active_conns")
	if q.has("conn_max_idle_time") {
		opts.ConnMaxIdleTime = q.duration("conn_max_idle_time")
	} else {
		opts.ConnMaxIdleTime = q.duration("idle_timeout")
	}
	if q.has("conn_max_lifetime") {
		opts.ConnMaxLifetime = q.duration("conn_max_lifetime")
	} else {
		opts.ConnMaxLifetime = q.duration("max_conn_age")
	}
	if q.err != nil {
		return nil, q.err
	}

	// any parameters left?
	if r := q.remaining(); len(r) > 0 {
		return nil, fmt.Errorf("redis: unexpected option: %s", strings.Join(r, ", "))
	}

	return &opts, nil
}

func NewRedisSentinel(opt *SentinelOptions) (*Redis, error) {
	failoverOptions, err := opt.failoverOptions()
	if err != nil {
		return nil, err
	}

	client := redis.NewFailoverClient(failoverOptions)
	err = client.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	if err = redisotel.InstrumentTracing(client); err != nil {
		return nil, err
	}

	if err = redisotel.InstrumentMetrics(client); err != nil {
		return nil, err
	}

	opts := options.NewOptions().SetSerialize(options.SerializeJSON)
	if opt.Exp > 0 {
		opts.SetExpire(time.Duration(opt.Exp) * time.Second)
	}

	cl := &Redis{
		Client:  client,
		options: opts,
	}

	return cl, nil
}
