package redis

import (
	"context"
	"errors"
	"github.com/dedetia/golib/storage/redis/options"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"reflect"
	"time"
)

type (
	Options struct {
		URI string
		DB  int
		Exp int
	}

	Redis struct {
		*redis.Client
		options *options.Options
	}

	DataSet struct {
		Key   string
		Value any
	}
)

func NewRedis(opt *Options) (*Redis, error) {
	redisOptions, err := redis.ParseURL(opt.URI)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(redisOptions)
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

func (r *Redis) GetInt(ctx context.Context, key string) (int, error) {
	return r.Client.Get(ctx, key).Int()
}

func (r *Redis) GetString(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// GetObject get cache object
func (r *Redis) GetObject(ctx context.Context, key string, value interface{}, opts ...*options.Options) error {
	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	opt := r.options.Merge(opts...)

	return opt.Serialize.Unmarshal(data, value)
}

func (r *Redis) RemainingTime(ctx context.Context, key string) int {
	return int(r.Client.TTL(ctx, key).Val().Seconds())
}

func (r *Redis) Exists(ctx context.Context, key ...string) bool {
	return r.Client.Exists(ctx, key...).Val() > 0
}

// MGet get cache with multiple keys
func (r *Redis) MGet(ctx context.Context, keys []string, obj interface{}, opts ...*options.Options) error {
	data, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return err
	}

	return r.readSliceInterfaceToObj(data, obj, opts...)
}

func (r *Redis) readSliceInterfaceToObj(val []interface{}, obj interface{}, opts ...*options.Options) error {
	opt := r.options.Merge(opts...)

	rv := reflect.ValueOf(obj)

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	for _, v := range val {
		if v == nil {
			continue
		}
		var elem reflect.Value

		typ := rv.Type().Elem()
		elem = reflect.New(typ.Elem())
		err := opt.Serialize.Unmarshal([]byte(v.(string)), elem.Interface())
		if err != nil {
			return err
		}
		rv.Set(reflect.Append(rv, elem))
	}

	return nil
}

func (r *Redis) GetMembers(ctx context.Context, key string, obj interface{}, opts ...*options.Options) (int, error) {
	keys, err := r.Client.SMembers(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	totalMembers := len(keys)
	if totalMembers == 0 {
		return 0, errors.New("member not found")
	}

	return totalMembers, r.MGet(ctx, keys, obj, opts...)
}

// Set cache will use the default expiration if the expiration is not filled
func (r *Redis) Set(ctx context.Context, key string, value interface{}, opts ...*options.Options) error {
	opt := r.options.Merge(opts...)

	switch value.(type) {
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, []byte:
		return r.Client.Set(ctx, key, value, opt.Expire).Err()
	default:
		data, err := opt.Serialize.Marshal(value)
		if err != nil {
			return err
		}

		return r.Client.Set(ctx, key, data, opt.Expire).Err()
	}
}

func (r *Redis) Del(ctx context.Context, key ...string) error {
	return r.Client.Del(ctx, key...).Err()
}

func (r *Redis) deleteWithPattern(ctx context.Context, pattern string) error {
	iter := r.Client.Scan(ctx, 0, pattern, 0).Iterator()
	var localKeys []string

	for iter.Next(ctx) {
		localKeys = append(localKeys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(localKeys) > 0 {
		_, err := r.Client.Pipelined(ctx, func(pipeline redis.Pipeliner) error {
			pipeline.Del(ctx, localKeys...)
			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteWithPattern clear cache with a pattern
func (r *Redis) DeleteWithPattern(ctx context.Context, pattern string) error {
	return r.deleteWithPattern(ctx, pattern)
}

func (r *Redis) SetMultiple(ctx context.Context, data []*DataSet, opts ...*options.Options) error {
	opt := r.options.Merge(opts...)

	_, err := r.Client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, v := range data {
			vb, err := opt.Serialize.Marshal(v.Value)
			if err != nil {
				return err
			}
			pipe.Set(ctx, v.Key, vb, opt.Expire)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) SetMember(ctx context.Context, key string, data []*DataSet, opts ...*options.Options) error {
	opt := r.options.Merge(opts...)

	members := make([]string, 0)
	_, err := r.Client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, v := range data {
			switch v.Value.(type) {
			case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, []byte:
				pipe.Set(ctx, v.Key, v.Value, opt.Expire)
			default:
				vb, err := opt.Serialize.Marshal(v.Value)
				if err != nil {
					return err
				}
				pipe.Set(ctx, v.Key, vb, opt.Expire)
			}
			members = append(members, v.Key)
		}
		pipe.SAdd(ctx, key, members)
		pipe.Expire(ctx, key, opt.Expire)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
